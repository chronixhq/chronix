package execution

import (
	"context"
	"log/slog"
	"time"

	"chronix/internal/db/models"
)

type WebTaskStepTestResult struct {
	Order           int64             `json:"order"`
	Name            string            `json:"name"`
	Status          string            `json:"status"`
	RequestMethod   string            `json:"requestMethod"`
	RequestURL      string            `json:"requestUrl"`
	RequestHeaders  map[string]string `json:"requestHeaders"`
	RequestBody     string            `json:"requestBody"`
	ResponseStatus  int               `json:"responseStatus"`
	ResponseHeaders map[string]string `json:"responseHeaders"`
	ResponseBody    string            `json:"responseBody"`
	LatencyMs       int64             `json:"latencyMs"`
	ExpectationOK   bool              `json:"expectationOk"`
	ExpectationMsg  string            `json:"expectationMsg"`
	ExecutionError  string            `json:"executionError,omitempty"`
	CapturedVars    map[string]any    `json:"capturedVars,omitempty"`
}

func TestWebTaskAction(ctx context.Context, steps []models.WebtaskActionStep, conn *models.WebtaskConnection, vars map[string]any) ([]WebTaskStepTestResult, error) {
	results := make([]WebTaskStepTestResult, 0, len(steps))

	slog.Info("Testing Web Task Action", "steps_count", len(steps), "conn_id", *conn.ID)

	for i, st := range steps {
		slog.Info("Running test step", "step", i+1, "name", st.Name, "url", st.URL, "method", st.Method)
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		res := WebTaskStepTestResult{
			Order: int64(i + 1),
			Name:  st.Name,
		}

		to := defaultStepTimeout
		if st.TimeoutSeconds != nil && *st.TimeoutSeconds > 0 {
			to = time.Duration(*st.TimeoutSeconds) * time.Second
		}

		result, runErr := runWebTask(ctx, conn, &st, vars, to)

		if runErr != nil {
			res.Status = "error"
			res.ExecutionError = runErr.Error()
		} else {
			res.Status = "success"
			res.RequestMethod = result.RequestMethod
			res.RequestURL = result.RequestURL
			res.RequestHeaders = result.RequestHeaders
			res.RequestBody = string(result.RequestBody)
			res.ResponseStatus = result.StatusCode
			res.ResponseHeaders = make(map[string]string)
			for k, v := range result.ResponseHeaders {
				if len(v) > 0 {
					res.ResponseHeaders[k] = v[0]
				}
			}
			res.ResponseBody = string(result.ResponseBody)
			res.LatencyMs = result.Latency.Milliseconds()

			if st.Expectation != nil {
				ok, msg, _ := evaluateWebTaskExpectation(*st.Expectation, result, vars)
				res.ExpectationOK = ok
				res.ExpectationMsg = msg
				if !ok {
					res.Status = "error"
				}
			} else {
				res.ExpectationOK = true
				res.ExpectationMsg = "ok"
			}
		}

		if res.Status == "success" {
			newVars := captureWebTaskVariables(st.ResponseCapture, result, vars)
			if len(newVars) > 0 {
				res.CapturedVars = newVars
				for k, v := range newVars {
					vars[k] = v
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
