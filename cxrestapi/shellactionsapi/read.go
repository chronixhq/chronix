package shellactionsapi

import (
	"chronix/cxrestapi/apiutil"
	"chronix/internal/db"
	"errors"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func listShellActions(c *gin.Context) {
	items, err := db.Action.Where(db.Action.ActionType.Eq("shell")).Order(db.Action.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list shell actions", err.Error())
		return
	}

	ids := make([]int64, 0, len(items))
	for _, a := range items {
		if a.ID != nil {
			ids = append(ids, *a.ID)
		}
	}

	stepsByAction := map[int64][]gin.H{}
	if len(ids) > 0 {
		rows, err := db.ShellActionStep.Where(db.ShellActionStep.ActionID.In(ids...)).Order(db.ShellActionStep.StepOrder.Asc()).Find()
		if err == nil {
			for _, s := range rows {
				env := map[string]any{}
				if s.EnvJSON != nil {
					for k, v := range *s.EnvJSON {
						env[k] = v
					}
				}
				stepsByAction[s.ActionID] = append(stepsByAction[s.ActionID], gin.H{
					"id":                    s.ID,
					"order":                 s.StepOrder,
					"name":                  s.Name,
					"runMode":               s.RunMode,
					"command":               s.Command,
					"scriptText":            s.ScriptText,
					"shellPath":             s.ShellPath,
					"workingDir":            s.WorkingDir,
					"timeoutSeconds":        s.TimeoutSeconds,
					"env":                   env,
					"outputCaptureMaxBytes": s.OutputCaptureMaxBytes,
					"outputTruncation":      s.OutputTruncation,
					"expectation":           s.Expectation,
					"outputCapture":         s.OutputCapture,
					"onFailure":             s.OnFailure,
				})
			}
		}
	}

	resp := make([]gin.H, 0, len(items))
	for _, a := range items {
		steps := stepsByAction[pickID(a.ID)]
		resp = append(resp, gin.H{
			"id":          a.ID,
			"name":        a.Name,
			"description": a.Description,
			"notes":       a.Notes,
			"enabled":     a.Enabled,
			"suspended":   a.Suspended,
			"createdAt":   a.CreatedAt,
			"updatedAt":   a.UpdatedAt,
			"steps":       steps,
		})
	}
	restresponse.RestSuccess(c, resp)
}

func getShellAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	a, err := db.Action.Where(db.Action.ID.Eq(id), db.Action.ActionType.Eq("shell")).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}

	steps, err := db.ShellActionStep.Where(db.ShellActionStep.ActionID.Eq(id)).Order(db.ShellActionStep.StepOrder.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load steps", err.Error())
		return
	}

	mapped := make([]gin.H, 0, len(steps))
	for _, s := range steps {
		env := map[string]any{}
		if s.EnvJSON != nil {
			for k, v := range *s.EnvJSON {
				env[k] = v
			}
		}
		mapped = append(mapped, gin.H{
			"id":                    s.ID,
			"order":                 s.StepOrder,
			"name":                  s.Name,
			"runMode":               s.RunMode,
			"command":               s.Command,
			"scriptText":            s.ScriptText,
			"shellPath":             s.ShellPath,
			"workingDir":            s.WorkingDir,
			"timeoutSeconds":        s.TimeoutSeconds,
			"env":                   env,
			"outputCaptureMaxBytes": s.OutputCaptureMaxBytes,
			"outputTruncation":      s.OutputTruncation,
			"expectation":           s.Expectation,
			"outputCapture":         s.OutputCapture,
			"onFailure":             s.OnFailure,
		})
	}

	restresponse.RestSuccess(c, gin.H{
		"id":          a.ID,
		"name":        a.Name,
		"description": a.Description,
		"notes":       a.Notes,
		"enabled":     a.Enabled,
		"suspended":   a.Suspended,
		"createdAt":   a.CreatedAt,
		"updatedAt":   a.UpdatedAt,
		"steps":       mapped,
	})
}
