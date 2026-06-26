package execution

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"chronix/internal/db"
	"chronix/internal/db/models"
	progresspkg "chronix/internal/progress"
	"chronix/internal/secret"
	"chronix/internal/shellrun"
)

// executeShellJob runs a shell-type action using shell_action_steps and a shell_connection.
func executeShellJob(ctx context.Context, logger *slog.Logger, job *models.Job, action *models.Action, runUID string, vars map[string]any) (string, string, error) {
	stepPtrs, err := db.ShellActionStep.Where(db.ShellActionStep.ActionID.Eq(*action.ID)).Order(db.ShellActionStep.StepOrder.Asc()).Find()
	if err != nil {
		logger.Error("load shell steps", "error", err)
		return "error", "failed to load shell action steps", err
	}
	steps := make([]models.ShellActionStep, 0, len(stepPtrs))
	for _, sp := range stepPtrs {
		if sp != nil {
			steps = append(steps, *sp)
		}
	}
	if len(steps) == 0 {
		return "success", "no steps", nil
	}
	if job.ShellConnectionID == nil || *job.ShellConnectionID == 0 {
		return "error", "missing shell connection id", errors.New("missing shell_connection_id")
	}
	connRow, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(*job.ShellConnectionID)).First()
	if err != nil {
		logger.Error("load shell connection", "error", err)
		return "error", "failed to load shell connection", err
	}
	conn := *connRow
	if conn.Suspended {
		return "error", "shell connection is suspended", errors.New("connection suspended")
	}
	conn.SSHPassword, _ = secret.DecryptPtr(conn.SSHPassword)
	conn.SSHPrivateKey, _ = secret.DecryptPtr(conn.SSHPrivateKey)
	conn.SSHKeyPass, _ = secret.DecryptPtr(conn.SSHKeyPass)
	conn.SudoPassword, _ = secret.DecryptPtr(conn.SudoPassword)

	if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr != nil && jr.ID != nil {
		for i, st := range steps {
			order := int64(i + 1)
			if cnt, _ := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID), db.JobRunStep.StepOrder.Eq(order)).Count(); cnt == 0 {
				stepRow := &models.JobRunStep{RunID: *jr.ID, StepOrder: &order, StepName: &st.Name, Status: "queued"}
				if st.TimeoutSeconds != nil {
					stepRow.TimeoutSeconds = st.TimeoutSeconds
				}
				if st.Command != nil {
					s := substituteVars(*st.Command, vars)
					stepRow.CommandText = &s
				}
				if st.ScriptText != nil {
					s := substituteVars(*st.ScriptText, vars)
					stepRow.ScriptText = &s
				}
				if st.ShellPath != nil {
					s := substituteVars(*st.ShellPath, vars)
					stepRow.ShellPath = &s
				}
				if st.WorkingDir != nil {
					s := substituteVars(*st.WorkingDir, vars)
					stepRow.WorkingDir = &s
				}
				if err := db.JobRunStep.Create(stepRow); err != nil {
					logger.Warn("precreate shell step", "error", err, "order", order, "step", st.Name)
				}
			}
		}
	}

	for idx, st := range steps {
		select {
		case <-ctx.Done():
			progresspkg.OnStepFinished(*job.ID, runUID, idx, st.Name, "canceled", "canceled", nil)
			return "canceled", "canceled", ctx.Err()
		default:
		}
		progresspkg.OnStepStarted(*job.ID, runUID, idx, st.Name)
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
		mode := strings.ToLower(pickStr(&conn.Mode))
		shellPath := pickStr(st.ShellPath)
		if shellPath != "" {
			shellPath = substituteVars(shellPath, vars)
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
		}
		var subScript *string
		if st.ScriptText != nil {
			s := substituteVars(*st.ScriptText, vars)
			subScript = &s
		}

		subWorkingDir := st.WorkingDir
		if st.WorkingDir != nil {
			s := substituteVars(*st.WorkingDir, vars)
			subWorkingDir = &s
		}

		if conn.AgentUUID != nil && strings.TrimSpace(*conn.AgentUUID) != "" {
			agentID := strings.TrimSpace(*conn.AgentUUID)
			var sshCfg *shellrun.SSHConfig
			if mode == "ssh" {
				pVal := int(pickInt64(conn.Port, 22))
				sshCfg = &shellrun.SSHConfig{
					Mode:          "ssh",
					Host:          conn.Host,
					Port:          &pVal,
					Username:      conn.SSHUsername,
					AuthMethod:    conn.AuthMethod,
					Password:      conn.SSHPassword,
					PrivateKey:    conn.SSHPrivateKey,
					KeyPassphrase: conn.SSHKeyPass,
					SudoPassword:  conn.SudoPassword,
				}
			}
			exitCode, stdout, stderr, runErr = shellrun.RunAgent(stepCtx, agentID, shellPath, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env, sshCfg, pickBool(conn.Sudo), conn.RunAsUser, conn.SudoPassword)
		} else if mode == "localhost" || mode == "" {
			sp := shellPath
			if sp == "" {
				sp = "/bin/sh"
			}
			exitCode, stdout, stderr, runErr = shellrun.RunLocal(stepCtx, sp, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env, pickBool(conn.Sudo), conn.RunAsUser, conn.SudoPassword)
		} else {
			sp := shellPath
			if sp == "" {
				sp = "/bin/sh"
			}
			exitCode, stdout, stderr, runErr = shellrun.RunSSH(stepCtx, &conn, sp, pickStr(st.RunMode), subCommand, subScript, subWorkingDir, env)
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
		outSeg, outTrunc, outTotal := shellrun.TruncateBytes(stdout, int(capBytes), trunc)
		errSeg, errTrunc, errTotal := shellrun.TruncateBytes(stderr, int(capBytes), trunc)
		if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr.ID != nil {
			order := int64(idx + 1)
			if stp, err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID), db.JobRunStep.StepOrder.Eq(order)).First(); err == nil && stp != nil && stp.ID != nil {
				outText := string(outSeg)
				errText := string(errSeg)
				ob := int64(len(outSeg))
				eb := int64(len(errSeg))
				ot := int64(outTotal)
				et := int64(errTotal)
				ioRow := &models.JobRunShellIo{StepID: *stp.ID, StdoutText: &outText, StderrText: &errText, StdoutTruncated: outTrunc, StderrTruncated: errTrunc, StdoutBytes: &ob, StderrBytes: &eb, StdoutTotalBytes: &ot, StderrTotalBytes: &et}
				_ = db.JobRunShellIo.Create(ioRow)
				ec := int64(exitCode)
				_, _ = db.JobRunStep.Where(db.JobRunStep.ID.Eq(*stp.ID)).UpdateSimple(db.JobRunStep.ExitCode.Value(ec))
			}
		}
		expectOK := true
		expectMsg := ""
		var expectMeta map[string]any
		if st.Expectation != nil {
			expectOK, expectMsg, expectMeta = evaluateShellExpectation(*st.Expectation, exitCode, outSeg, errSeg, outTrunc, errTrunc, trunc, vars)
		} else if exitCode != 0 {
			expectOK = false
			expectMsg = fmt.Sprintf("exit %d", exitCode)
		}

		fields := map[string]any{}
		for k, v := range expectMeta {
			fields[k] = v
		}
		if subCommand != nil {
			fields["command_text"] = *subCommand
		}
		if subScript != nil {
			fields["script_text"] = *subScript
		}
		if shellPath != "" {
			fields["shell_path"] = shellPath
		}
		if subWorkingDir != nil {
			fields["working_dir"] = *subWorkingDir
		}

		if st.OutputCapture != nil {
			newVars := captureShellVariables(st.OutputCapture, outSeg, errSeg, vars)
			if len(newVars) > 0 {
				fields["captured_vars"] = newVars
				fields["result_lines"] = []map[string]any{newVars}
				for k, v := range newVars {
					vars[k] = v
				}
			}
		}

		stepFailed := false
		failureDetail := ""
		if runErr != nil {
			stepFailed = true
			failureDetail = runErr.Error()
		} else if !expectOK {
			stepFailed = true
			failureDetail = expectMsg
		}

		if stepFailed {
			onFailure := "exit"
			if st.OnFailure != nil && *st.OnFailure != "" {
				onFailure = *st.OnFailure
			}
			fields["error_message"] = failureDetail

			if onFailure == "exit" {
				progresspkg.OnStepFinished(*job.ID, runUID, idx, st.Name, "error", failureDetail, fields)
				return "error", failureDetail, runErr
			}
			progresspkg.OnStepFinished(*job.ID, runUID, idx, st.Name, "warning", failureDetail+" (continuing)", fields)
			continue
		}

		progresspkg.OnStepFinished(*job.ID, runUID, idx, st.Name, "success", "ok", fields)
	}
	return "success", "ok", nil
}

func pickStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func pickInt64(p *int64, def int64) int64 {
	if p == nil {
		return def
	}
	return *p
}

func pickBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}
