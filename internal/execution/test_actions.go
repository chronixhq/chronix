package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/internal/shellrun"
	"chronix/internal/sqlrunner"
)

type DatabaseStepTestResult struct {
	Order          int64            `json:"order"`
	Name           string           `json:"name"`
	Status         string           `json:"status"` // success | error | warning | canceled
	ExecutedCode   string           `json:"executedCode"`
	ExecutedArgs   []any            `json:"executedArgs,omitempty"`
	RowsCount      int              `json:"rowsCount"`
	RowsAffected   int64            `json:"rowsAffected"`
	ResultLines    []map[string]any `json:"resultLines,omitempty"`
	ExpectationOK  bool             `json:"expectationOk"`
	ExpectationMsg string           `json:"expectationMsg"`
	ExecutionError string           `json:"executionError,omitempty"`
	CapturedVars   map[string]any   `json:"capturedVars,omitempty"`
}

type ShellStepTestResult struct {
	Order           int64          `json:"order"`
	Name            string         `json:"name"`
	Status          string         `json:"status"` // success | error | warning | canceled
	ExecutedCode    string         `json:"executedCode"`
	ExitCode        int            `json:"exitCode"`
	Stdout          string         `json:"stdout"`
	Stderr          string         `json:"stderr"`
	StdoutTruncated bool           `json:"stdoutTruncated"`
	StderrTruncated bool           `json:"stderrTruncated"`
	ExpectationOK   bool           `json:"expectationOk"`
	ExpectationMsg  string         `json:"expectationMsg"`
	ExecutionError  string         `json:"executionError,omitempty"`
	CapturedVars    map[string]any `json:"capturedVars,omitempty"`
}

func TestDatabaseAction(ctx context.Context, steps []models.ActionStep, conn *models.DbConnection, varMap map[string]any) ([]DatabaseStepTestResult, error) {
	results := make([]DatabaseStepTestResult, 0, len(steps))

	driver := strings.ToLower(conn.Driver)
	dsn := ""
	if conn.Dsn != nil {
		dsn, _ = secret.Decrypt(*conn.Dsn)
	}

	var runner sqlrunner.SQLRunner
	if conn.AgentUUID != nil && strings.TrimSpace(*conn.AgentUUID) != "" {
		runner = sqlrunner.AgentRunner{AgentID: strings.TrimSpace(*conn.AgentUUID)}
	} else {
		runner = sqlrunner.LocalRunner{}
	}

	sessionErr := runner.RunInSession(ctx, driver, dsn, func(s sqlrunner.SessionRunner) error {
		for i, step := range steps {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			res := DatabaseStepTestResult{
				Order: int64(i + 1),
				Name:  step.Name,
			}

			to := defaultStepTimeout
			if step.TimeoutSeconds != nil && *step.TimeoutSeconds > 0 {
				to = time.Duration(*step.TimeoutSeconds) * time.Second
			}

			stepCtx, cancel := context.WithTimeout(ctx, to)
			var runErr error
			var rowsCount int
			var rowsAffected int64
			var resultLines []map[string]any

			q, args, bindErr := bindParams(driver, step.SqlText, varMap)
			res.ExecutedCode = q
			res.ExecutedArgs = args
			if bindErr != nil {
				runErr = bindErr
			} else {
				if isSelectLike(step.SqlText) {
					rowsCount, resultLines, runErr = s.Query(stepCtx, q, args, resultRowCap)
				} else {
					rowsAffected, runErr = s.Exec(stepCtx, q, args)
				}
			}
			cancel()

			if runErr != nil {
				res.Status = "error"
				res.ExecutionError = runErr.Error()
			} else {
				res.RowsCount = rowsCount
				res.RowsAffected = rowsAffected
				res.ResultLines = resultLines

				if step.Expectation != nil {
					ok, msg, _, evalErr := evaluateExpectation(*step.Expectation, rowsCount, rowsAffected, resultLines, varMap)
					if evalErr != nil {
						res.Status = "error"
						res.ExecutionError = fmt.Sprintf("expectation evaluation error: %v", evalErr)
					} else {
						res.ExpectationOK = ok
						res.ExpectationMsg = msg
						if ok {
							res.Status = "success"
							res.ExpectationMsg = "ok"
						} else {
							res.Status = "error"
						}
					}
				} else {
					res.Status = "success"
					res.ExpectationOK = true
					res.ExpectationMsg = "ok"
				}

				if res.Status == "success" && step.OutputCapture != nil {
					newVars := captureDatabaseVariables(step.OutputCapture, rowsCount, resultLines, varMap)
					if len(newVars) > 0 {
						res.CapturedVars = newVars
						for k, v := range newVars {
							varMap[k] = v
						}
					}
				}
			}

			results = append(results, res)
			if res.Status == "error" {
				onFailure := "exit"
				if step.OnFailure != nil && *step.OnFailure != "" {
					onFailure = *step.OnFailure
				}
				if onFailure == "exit" {
					break
				}
			}
		}
		return nil
	})

	return results, sessionErr
}

func TestShellAction(ctx context.Context, steps []models.ShellActionStep, conn *models.ShellConnection, vars map[string]any) ([]ShellStepTestResult, error) {
	results := make([]ShellStepTestResult, 0, len(steps))

	sc := *conn
	sc.SSHPassword, _ = secret.DecryptPtr(sc.SSHPassword)
	sc.SSHPrivateKey, _ = secret.DecryptPtr(sc.SSHPrivateKey)
	sc.SSHKeyPass, _ = secret.DecryptPtr(sc.SSHKeyPass)
	sc.SudoPassword, _ = secret.DecryptPtr(sc.SudoPassword)

	for i, st := range steps {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		res := ShellStepTestResult{
			Order: int64(i + 1),
			Name:  st.Name,
		}

		to := defaultStepTimeout
		if st.TimeoutSeconds != nil && *st.TimeoutSeconds > 0 {
			to = time.Duration(*st.TimeoutSeconds) * time.Second
		}

		stepCtx, cancel := context.WithTimeout(ctx, to)

		env := map[string]string{}
		if st.EnvJSON != nil {
			for k, v := range *st.EnvJSON {
				if s, ok := v.(string); ok {
					env[k] = substituteVars(s, vars)
				}
			}
		}

		mode := strings.ToLower(pickStr(&sc.Mode))
		shellPath := pickStr(st.ShellPath)
		if shellPath != "" {
			shellPath = substituteVars(shellPath, vars)
		} else {
			if sc.AgentUUID != nil && strings.TrimSpace(*sc.AgentUUID) != "" && (mode == "localhost" || mode == "") {
				shellPath = ""
			} else {
				shellPath = "/bin/sh"
			}
		}

		var (
			exitCode int
			stdout   []byte
			stderr   []byte
			runErr   error
		)

		var subCommand *string
		if st.Command != nil {
			s := substituteVars(*st.Command, vars)
			subCommand = &s
			res.ExecutedCode = s
		}
		var subScript *string
		if st.ScriptText != nil {
			s := substituteVars(*st.ScriptText, vars)
			subScript = &s
			res.ExecutedCode = s
		}

		subWorkingDir := st.WorkingDir
		if st.WorkingDir != nil {
			s := substituteVars(*st.WorkingDir, vars)
			subWorkingDir = &s
		}

		if sc.AgentUUID != nil && strings.TrimSpace(*sc.AgentUUID) != "" {
			agentID := strings.TrimSpace(*sc.AgentUUID)
			var sshCfg *shellrun.SSHConfig
			if mode == "ssh" {
				pVal := int(pickInt64(sc.Port, 22))
				sshCfg = &shellrun.SSHConfig{
					Mode:          "ssh",
					Host:          sc.Host,
					Port:          &pVal,
					Username:      sc.SSHUsername,
					AuthMethod:    sc.AuthMethod,
					Password:      sc.SSHPassword,
					PrivateKey:    sc.SSHPrivateKey,
					KeyPassphrase: sc.SSHKeyPass,
					SudoPassword:  sc.SudoPassword,
				}
			}
			exitCode, stdout, stderr, runErr = shellrun.RunAgent(stepCtx, agentID, shellPath, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env, sshCfg, pickBool(sc.Sudo), sc.RunAsUser, sc.SudoPassword)
		} else {
			if mode == "localhost" || mode == "" {
				exitCode, stdout, stderr, runErr = shellrun.RunLocal(stepCtx, shellPath, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env, pickBool(sc.Sudo), sc.RunAsUser, sc.SudoPassword)
			} else {
				exitCode, stdout, stderr, runErr = shellrun.RunSSH(stepCtx, &sc, shellPath, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env)
			}
		}
		cancel()

		capBytes := int64(65536)
		trunc := "tail"
		if st.OutputCaptureMaxBytes != nil && *st.OutputCaptureMaxBytes > 0 {
			capBytes = *st.OutputCaptureMaxBytes
		}
		if st.OutputTruncation != nil && *st.OutputTruncation != "" {
			trunc = *st.OutputTruncation
		}
		outSeg, outTrunc, _ := shellrun.TruncateBytes(stdout, int(capBytes), trunc)
		errSeg, errTrunc, _ := shellrun.TruncateBytes(stderr, int(capBytes), trunc)

		res.ExitCode = exitCode
		res.Stdout = string(outSeg)
		res.Stderr = string(errSeg)
		res.StdoutTruncated = outTrunc
		res.StderrTruncated = errTrunc

		if runErr != nil {
			res.Status = "error"
			res.ExecutionError = runErr.Error()
		} else {
			if st.Expectation != nil {
				ok, msg, _ := evaluateShellExpectation(*st.Expectation, exitCode, outSeg, errSeg, outTrunc, errTrunc, trunc, vars)
				res.ExpectationOK = ok
				res.ExpectationMsg = msg
				if ok {
					res.Status = "success"
					res.ExpectationMsg = "ok"
				} else {
					res.Status = "error"
				}
			} else {
				if exitCode == 0 {
					res.Status = "success"
					res.ExpectationOK = true
					res.ExpectationMsg = "ok"
				} else {
					res.Status = "error"
					res.ExpectationOK = false
					res.ExpectationMsg = fmt.Sprintf("exit %d", exitCode)
				}
			}

			if res.Status == "success" && st.OutputCapture != nil {
				newVars := captureShellVariables(st.OutputCapture, outSeg, errSeg, vars)
				if len(newVars) > 0 {
					res.CapturedVars = newVars
					for k, v := range newVars {
						vars[k] = v
					}
				}
			}
		}

		results = append(results, res)
		if res.Status == "error" {
			onFailure := "exit"
			if st.OnFailure != nil && *st.OnFailure != "" {
				onFailure = *st.OnFailure
			}
			if onFailure == "exit" {
				break
			}
		}
	}

	return results, nil
}
