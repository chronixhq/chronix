package progress

import (
	"chronix/internal/db"

	"github.com/dan-sherwin/go-utilities"
)

func AggregateRunOutput(runUID string) map[string]any {
	out := map[string]any{}
	jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).First()
	if err != nil || jr == nil || jr.ID == nil {
		return out
	}

	steps, err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID)).Order(db.JobRunStep.StepOrder.Asc()).Find()
	if err != nil {
		return out
	}

	stepOutputs := make([]any, 0, len(steps))
	for _, s := range steps {
		stepData := map[string]any{
			"name":           utilities.PtrVal(s.StepName),
			"status":         s.Status,
			"rows_count":     utilities.PtrVal(s.RowsCount),
			"rows_affected":  utilities.PtrVal(s.RowsAffected),
			"expect_ok":      utilities.PtrVal(s.ExpectOk),
			"expect_message": utilities.PtrVal(s.ExpectMessage),
			"error_message":  utilities.PtrVal(s.ErrorMessage),
		}
		if s.Details != nil {
			if rl, ok := (*s.Details)["result_lines"]; ok {
				stepData["result_lines"] = rl
			}
		}
		if s.ID != nil {
			if shellIO, err := db.JobRunShellIo.Where(db.JobRunShellIo.StepID.Eq(*s.ID)).First(); err == nil && shellIO != nil {
				if shellIO.StdoutText != nil {
					stepData["stdout"] = *shellIO.StdoutText
				}
				if shellIO.StderrText != nil {
					stepData["stderr"] = *shellIO.StderrText
				}
			}
			if webIO, err := db.JobRunWebtaskIo.Where(db.JobRunWebtaskIo.StepID.Eq(*s.ID)).First(); err == nil && webIO != nil {
				if webIO.ResponseStatus != nil {
					stepData["response_status"] = *webIO.ResponseStatus
				}
				if webIO.ResponseBody != nil {
					stepData["response_body"] = *webIO.ResponseBody
				}
			}
		}
		stepOutputs = append(stepOutputs, stepData)
	}
	out["steps"] = stepOutputs
	return out
}
