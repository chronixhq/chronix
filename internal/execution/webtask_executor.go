package execution

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"time"

	"chronix/internal/db"
	"chronix/internal/db/models"
	progresspkg "chronix/internal/progress"
	"chronix/internal/secret"
	"chronix/internal/webtaskrun"

	"github.com/dan-sherwin/go-utilities"
	"gorm.io/datatypes"
)

func executeWebTaskJob(ctx context.Context, logger *slog.Logger, job *models.Job, action *models.Action, runUID string, vars map[string]any) (string, string, error) {
	stepPtrs, err := db.WebtaskActionStep.Where(db.WebtaskActionStep.ActionID.Eq(*action.ID)).Order(db.WebtaskActionStep.StepOrder.Asc()).Find()
	if err != nil {
		logger.Error("load webtask steps", "error", err)
		return "error", "failed to load webtask action steps", err
	}
	steps := make([]models.WebtaskActionStep, 0, len(stepPtrs))
	for _, sp := range stepPtrs {
		if sp != nil {
			steps = append(steps, *sp)
		}
	}
	if len(steps) == 0 {
		return "success", "no steps", nil
	}

	if job.WebtaskConnectionID == nil || *job.WebtaskConnectionID == 0 {
		return "error", "missing webtask connection id", errors.New("missing webtask_connection_id")
	}
	connRow, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(*job.WebtaskConnectionID)).First()
	if err != nil {
		logger.Error("load webtask connection", "error", err)
		return "error", "failed to load webtask connection", err
	}
	conn := *connRow
	if conn.Suspended != nil && *conn.Suspended {
		return "error", "webtask connection is suspended", errors.New("connection suspended")
	}

	if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr != nil && jr.ID != nil {
		for i, st := range steps {
			order := int64(i + 1)
			if cnt, _ := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID), db.JobRunStep.StepOrder.Eq(order)).Count(); cnt == 0 {
				stepRow := &models.JobRunStep{RunID: *jr.ID, StepOrder: &order, StepName: &st.Name, Status: "queued"}
				if st.TimeoutSeconds != nil {
					stepRow.TimeoutSeconds = st.TimeoutSeconds
				}
				_ = db.JobRunStep.Create(stepRow)
			}
		}
	}

	var totalRowsAffected int64
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

		result, runErr := runWebTask(ctx, &conn, &st, vars, to)

		status := "success"
		msg := "ok"
		expectOk := true
		var fields map[string]any

		if runErr != nil {
			status = "error"
			msg = runErr.Error()
			expectOk = false
		} else if st.Expectation != nil {
			expectOk, msg, fields = evaluateWebTaskExpectation(*st.Expectation, result, vars)
			if fields != nil {
				if ra, ok := fields["rows_affected"].(int64); ok {
					totalRowsAffected += ra
				}
			}
			if !expectOk {
				status = "error"
			}
		}

		if fields == nil {
			fields = map[string]any{}
		}

		newVars := captureWebTaskVariables(st.ResponseCapture, result, vars)
		if len(newVars) > 0 {
			fields["captured_vars"] = newVars
			fields["result_lines"] = []map[string]any{newVars}
			for k, v := range newVars {
				vars[k] = v
			}
		}

		if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr != nil {
			if jrs, err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID), db.JobRunStep.StepOrder.Eq(int64(idx+1))).First(); err == nil && jrs != nil {
				ioRow := &models.JobRunWebtaskIo{
					StepID:         *jrs.ID,
					RequestURL:     &result.RequestURL,
					RequestMethod:  &result.RequestMethod,
					RequestBody:    utilities.Ptr(string(result.RequestBody)),
					ResponseStatus: utilities.Ptr(int64(result.StatusCode)),
					ResponseBody:   utilities.Ptr(string(result.ResponseBody)),
					LatencyMs:      utilities.Ptr(result.Latency.Milliseconds()),
				}
				ioRow.RequestHeaders = toJSONMap(result.RequestHeaders)
				ioRow.ResponseHeaders = toJSONMap(result.ResponseHeaders)
				_ = db.JobRunWebtaskIo.Create(ioRow)

				jrs.Status = status
				jrs.ExpectOk = &expectOk
				jrs.ExpectMessage = &msg
				jrs.FinishedAt = utilities.Ptr(time.Now())
				_ = db.JobRunStep.Save(jrs)
			}
		}

		progresspkg.OnStepFinished(*job.ID, runUID, idx, st.Name, status, msg, fields)

		if status == "error" {
			onFailure := "exit"
			if st.OnFailure != nil && *st.OnFailure != "" {
				onFailure = *st.OnFailure
			}
			if onFailure == "exit" {
				return "error", msg, nil
			}
		}
	}

	if totalRowsAffected > 0 {
		_, _ = db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).UpdateSimple(db.JobRun.RowsAffected.Value(totalRowsAffected))
	}
	return "success", "completed", nil
}

func toJSONMap(m any) *datatypes.JSONMap {
	if m == nil {
		return nil
	}
	b, _ := json.Marshal(m)
	var jm datatypes.JSONMap
	_ = json.Unmarshal(b, &jm)
	return &jm
}

func runWebTask(ctx context.Context, conn *models.WebtaskConnection, st *models.WebtaskActionStep, vars map[string]any, timeout time.Duration) (*webtaskrun.WebTaskResult, error) {
	url := substituteVars(st.URL, vars)
	if conn.BaseURL != nil && *conn.BaseURL != "" && !strings.HasPrefix(strings.ToLower(url), "http") {
		url = strings.TrimSuffix(*conn.BaseURL, "/") + "/" + strings.TrimPrefix(url, "/")
	}

	slog.Info("Executing Web Task", "method", st.Method, "url", url)

	headers := map[string]string{}
	if st.Headers != nil {
		for k, v := range *st.Headers {
			if s, ok := v.(string); ok {
				headers[k] = substituteVars(s, vars)
			}
		}
	}

	authType := strings.ToLower(conn.AuthType)
	if authType != "none" && conn.AuthConfig != nil {
		config := *conn.AuthConfig
		switch authType {
		case "basic":
			user, _ := config["username"].(string)
			pass, _ := config["password"].(string)
			if pass != "" {
				pass, _ = secret.Decrypt(pass)
			}
			headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
		case "bearer":
			token, _ := config["token"].(string)
			if token != "" {
				token, _ = secret.Decrypt(token)
			}
			headers["Authorization"] = "Bearer " + token
		case "header":
			name, _ := config["header_name"].(string)
			value, _ := config["header_value"].(string)
			if value != "" {
				value, _ = secret.Decrypt(value)
			}
			if name != "" {
				headers[name] = value
			}
		}
	}

	var body []byte
	if st.Body != nil && *st.Body != "" {
		body = []byte(substituteVars(*st.Body, vars))
	}

	var runner webtaskrun.WebTaskRunner
	if conn.AgentUUID != nil && *conn.AgentUUID != "" {
		runner = &webtaskrun.AgentRunner{AgentID: *conn.AgentUUID}
	} else {
		runner = webtaskrun.NewLocalRunner()
	}

	return runner.Execute(ctx, st.Method, url, headers, body, timeout)
}
