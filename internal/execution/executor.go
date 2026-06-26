package execution

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"chronix/internal/db"
	"chronix/internal/db/models"
	progresspkg "chronix/internal/progress"
	"chronix/internal/secret"
	"chronix/internal/sqlrunner"

	// Register pgx stdlib driver for database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Default configuration (overridable via setters)
var (
	defaultStepTimeout = 30 * time.Second
	defaultRetryCount  = 1  // number of retries (not attempts). attempts = 1 + retries
	resultRowCap       = 50 // also used as the default max number of result lines recorded per step
)

// SetExecutorDefaults updates global executor settings like default timeouts and retry counts.
func SetExecutorDefaults(stepTimeout time.Duration, retryCount int, rowCap int) {
	if stepTimeout > 0 {
		defaultStepTimeout = stepTimeout
	}
	if retryCount >= 0 {
		defaultRetryCount = retryCount
	}
	if rowCap > 0 {
		resultRowCap = rowCap
	}
}

// ExecuteJob loads the job/action/steps/connection and runs sequentially with timeouts and simple retries.
func ExecuteJob(ctx context.Context, jobID int64, runID string) (string, string, error) {
	start := time.Now()
	logger := slog.With("component", "executor", "job_id", jobID, "run_id", runID)
	// Load job
	jobRow, err := db.Job.Where(db.Job.ID.Eq(jobID)).First()
	if err != nil {
		logger.Error("load job", "error", err)
		return "error", "failed to load job", err
	}
	job := *jobRow
	if job.Suspended {
		return "error", "job is suspended", errors.New("job suspended")
	}
	// Load action
	actionRow, err := db.Action.Where(db.Action.ID.Eq(job.ActionID)).First()
	if err != nil {
		logger.Error("load action", "error", err)
		return "error", "failed to load action", err
	}
	action := *actionRow
	if action.Suspended {
		return "error", "action is suspended", errors.New("action suspended")
	}
	// Load job variables
	vars, err := db.JobVariable.Where(db.JobVariable.JobID.Eq(*job.ID)).Find()
	if err != nil {
		logger.Error("load job vars", "error", err)
		return "error", "failed to load variables", err
	}
	varMap := map[string]any{}
	for _, v := range vars {
		varMap[v.Name] = v.Value
	}

	// Branch for shell vs database actions with minimal churn
	if strings.ToLower(action.ActionType) == "shell" || strings.ToLower(job.TargetKind) == "shell" {
		return executeShellJob(ctx, logger, &job, &action, runID, varMap)
	}
	if strings.ToLower(action.ActionType) == "webtask" || strings.ToLower(job.TargetKind) == "webtask" {
		return executeWebTaskJob(ctx, logger, &job, &action, runID, varMap)
	}
	// Database action path (existing behavior)
	// Load steps ordered
	stepPtrs, err := db.ActionStep.Where(db.ActionStep.ActionID.Eq(*action.ID)).Order(db.ActionStep.StepOrder.Asc()).Find()
	if err != nil {
		logger.Error("load steps", "error", err)
		return "error", "failed to load action steps", err
	}
	// de-pointer for convenience where needed
	steps := make([]models.ActionStep, 0, len(stepPtrs))
	for _, sp := range stepPtrs {
		if sp != nil {
			steps = append(steps, *sp)
		}
	}
	// Load connection
	if job.ConnectionID == nil {
		return "error", "missing connection id", errors.New("missing connection_id")
	}
	connRow, err := db.DbConnection.Where(db.DbConnection.ID.Eq(*job.ConnectionID)).First()
	if err != nil {
		logger.Error("load connection", "error", err)
		return "error", "failed to load connection", err
	}
	conn := *connRow
	if conn.Suspended != nil && *conn.Suspended {
		return "error", "database connection is suspended", errors.New("connection suspended")
	}

	// Prepare connection details and choose SQL runner (agent vs local)
	driver := strings.ToLower(conn.Driver)
	dsn := ""
	if conn.Dsn != nil {
		dsn, _ = secret.Decrypt(*conn.Dsn)
	}
	// Select runner based on agent assignment
	var runner sqlrunner.SQLRunner
	if conn.AgentUUID != nil && strings.TrimSpace(*conn.AgentUUID) != "" {
		runner = sqlrunner.AgentRunner{AgentID: strings.TrimSpace(*conn.AgentUUID)}
		logger.Info("using agent runner", "runner", "agent", "agent_id", strings.TrimSpace(*conn.AgentUUID))
	} else {
		runner = sqlrunner.LocalRunner{}
		logger.Info("using local runner", "runner", "local")
	}

	// Pre-create all run steps as 'queued' so the UI can display the full plan up-front
	// This makes RunDetail show all steps immediately and allows SSE/polling to update per-step status.
	// Best-effort: ignore errors for individual inserts to avoid blocking execution.
	if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runID)).Select(db.JobRun.ID).First(); err == nil && jr != nil && jr.ID != nil {
		// Load existing steps for this run to avoid duplicate Count() calls in the loop
		existing, _ := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID)).Select(db.JobRunStep.StepOrder).Find()
		existingOrders := make(map[int64]bool)
		for _, es := range existing {
			if es.StepOrder != nil {
				existingOrders[*es.StepOrder] = true
			}
		}

		for i, st := range steps {
			order := int64(i + 1)
			if !existingOrders[order] {
				stepRow := &models.JobRunStep{
					RunID:     *jr.ID,
					StepOrder: &order,
					StepName:  &st.Name,
					Status:    "queued",
				}
				// Optionally persist SQL/timeout so details are visible before start
				if st.SqlText != "" {
					subSQL, _, _ := bindParams(driver, st.SqlText, varMap)
					stepRow.SqlText = &subSQL
				}
				if st.TimeoutSeconds != nil {
					stepRow.TimeoutSeconds = st.TimeoutSeconds
				}
				if err := db.JobRunStep.Create(stepRow); err != nil {
					logger.Warn("precreate step row", "error", err, "order", order, "step", st.Name)
				}
			}
		}
	} else {
		logger.Warn("precreate steps: load run id failed", "error", err)
	}

	// Execute steps
	runResult, runMsg, runErr := "success", "completed", error(nil)
	sessionErr := runner.RunInSession(ctx, driver, dsn, func(s sqlrunner.SessionRunner) error {
		for idx, step := range steps {
			select {
			case <-ctx.Done():
				// User-initiated cancellation: mark step and run as "canceled" (not error)
				progresspkg.OnStepFinished(jobID, runID, idx, step.Name, "canceled", "canceled", nil)
				runResult, runMsg, runErr = "canceled", "canceled", ctx.Err()
				return ctx.Err()
			default:
			}
			// Announce start
			progresspkg.OnStepStarted(jobID, runID, idx, step.Name)
			// Determine timeout for step
			to := defaultStepTimeout
			if step.TimeoutSeconds != nil && *step.TimeoutSeconds > 0 {
				to = time.Duration(*step.TimeoutSeconds) * time.Second
			}
			attempts := 1 + defaultRetryCount
			var (
				rowsCount    int
				rowsAffected int64
				resultLines  []map[string]any
				lastErr      error
				executedCode string
				executedArgs []any
			)

			for attempt := 1; attempt <= attempts; attempt++ {
				ctxStep, cancel := context.WithTimeout(ctx, to)
				rowsCount = 0
				rowsAffected = 0
				resultLines = nil
				var stepRunErr error

				executedCode, executedArgs, stepRunErr = bindParams(driver, step.SqlText, varMap)
				if stepRunErr == nil {
					if isSelectLike(step.SqlText) {
						rowsCount, resultLines, stepRunErr = s.Query(ctxStep, executedCode, executedArgs, resultRowCap)
					} else {
						rowsAffected, stepRunErr = s.Exec(ctxStep, executedCode, executedArgs)
					}
				}
				cancel()

				if stepRunErr == nil {
					lastErr = nil
					break
				}

				// Failure path for execution
				lastErr = stepRunErr
				if attempt < attempts {
					// Backoff with jitter
					d := 500*time.Millisecond + time.Duration(rand.Intn(300))*time.Millisecond
					t := time.NewTimer(d)
					select {
					case <-ctx.Done():
						progresspkg.OnStepFinished(jobID, runID, idx, step.Name, "canceled", "canceled", nil)
						runResult, runMsg, runErr = "canceled", "canceled", ctx.Err()
						return ctx.Err()
					case <-t.C:
					}
					continue
				}
			}

			expectOK := true
			expectMsg := ""
			var expectMeta map[string]any
			var evalErr error

			if lastErr == nil {
				// Evaluate expectation if provided
				if step.Expectation != nil {
					expectOK, expectMsg, expectMeta, evalErr = evaluateExpectation(*step.Expectation, rowsCount, rowsAffected, resultLines, varMap)
				}
			}

			fields := map[string]any{
				"executed_code": executedCode,
				"executed_args": executedArgs,
			}
			if isSelectLike(step.SqlText) {
				fields["rows_count"] = rowsCount
				if len(resultLines) > 0 {
					fields["result_lines"] = resultLines
				}
			} else {
				fields["rows_affected"] = rowsAffected
			}
			for k, v := range expectMeta {
				fields[k] = v
			}
			if step.Expectation != nil {
				fields["expectation"] = *step.Expectation
			}

			stepFailed := false
			failureDetail := ""
			var stepFailureErr error

			if lastErr != nil {
				stepFailed = true
				failureDetail = lastErr.Error()
				stepFailureErr = lastErr
				fields["error_code"] = "sql_error"
			} else if evalErr != nil {
				stepFailed = true
				failureDetail = fmt.Sprintf("expectation error: %v", evalErr)
				stepFailureErr = evalErr
				fields["expect_ok"] = false
				fields["expect_message"] = failureDetail
				fields["error_code"] = "expectation_eval_error"
			} else if !expectOK {
				stepFailed = true
				failureDetail = expectMsg
				stepFailureErr = errors.New("assertion failed")
				fields["expect_ok"] = false
				fields["expect_message"] = expectMsg
			}

			if stepFailed {
				onFailure := "exit"
				if step.OnFailure != nil && *step.OnFailure != "" {
					onFailure = *step.OnFailure
				}
				fields["error_message"] = failureDetail

				if onFailure == "exit" {
					progresspkg.OnStepFinished(jobID, runID, idx, step.Name, "error", failureDetail, fields)
					runResult, runMsg, runErr = "error", failureDetail, stepFailureErr
					return nil // stop session
				}
				// continue
				progresspkg.OnStepFinished(jobID, runID, idx, step.Name, "warning", failureDetail+" (continuing)", fields)
				continue
			}

			// Success for this step
			if step.Expectation != nil {
				fields["expect_ok"] = true
				fields["expect_message"] = "ok"
			}

			// Capture variables
			if step.OutputCapture != nil {
				newVars := captureDatabaseVariables(step.OutputCapture, rowsCount, resultLines, varMap)
				for k, v := range newVars {
					varMap[k] = v
				}
			}

			progresspkg.OnStepFinished(jobID, runID, idx, step.Name, "success", "ok", fields)
		}
		return nil
	})

	if sessionErr != nil && runErr == nil {
		runResult, runMsg, runErr = "error", sessionErr.Error(), sessionErr
	}

	elapsed := time.Since(start)
	logger.Info("job finished", "status", runResult, "duration_ms", elapsed.Milliseconds())
	return runResult, runMsg, runErr
}
