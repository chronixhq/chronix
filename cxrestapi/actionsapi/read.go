package actionsapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"errors"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"chronix/cxrestapi/apiutil"
)

func listActions(c *gin.Context) {
	var items []*models.Action
	var err error
	if items, err = db.Action.Where(db.Action.ActionType.Eq("database")).Order(db.Action.Name.Asc()).Find(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list actions", err.Error())
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
		steps, err := db.ActionStep.
			Where(db.ActionStep.ActionID.In(ids...)).
			Order(db.ActionStep.StepOrder.Asc()).
			Find()
		if err == nil {
			for _, s := range steps {
				aid := s.ActionID
				stepsByAction[aid] = append(stepsByAction[aid], gin.H{
					"id":            s.ID,
					"order":         s.StepOrder,
					"name":          s.Name,
					"sqlText":       s.SqlText,
					"expectation":   s.Expectation,
					"outputCapture": s.OutputCapture,
					"onFailure":     s.OnFailure,
				})
			}
		}
	}

	resp := make([]gin.H, 0, len(items))
	for _, a := range items {
		var steps []gin.H
		if a.ID != nil {
			steps = stepsByAction[*a.ID]
		}
		resp = append(resp, gin.H{
			"id":          a.ID,
			"name":        a.Name,
			"dialect":     a.Dialect,
			"description": a.Description,
			"enabled":     a.Enabled,
			"suspended":   a.Suspended,
			"createdAt":   a.CreatedAt,
			"updatedAt":   a.UpdatedAt,
			"steps":       steps,
		})
	}
	restresponse.RestSuccess(c, resp)
}

func getAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	a, err := db.Action.Where(
		db.Action.ID.Eq(id),
		db.Action.ActionType.Eq("database"),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}

	steps, err := db.ActionStep.Where(db.ActionStep.ActionID.Eq(id)).Order(db.ActionStep.StepOrder.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load steps", err.Error())
		return
	}

	mapped := make([]gin.H, 0, len(steps))
	for _, s := range steps {
		mapped = append(mapped, gin.H{
			"id":             s.ID,
			"order":          s.StepOrder,
			"name":           s.Name,
			"sqlText":        s.SqlText,
			"timeoutSeconds": s.TimeoutSeconds,
			"expectation":    s.Expectation,
			"outputCapture":  s.OutputCapture,
			"onFailure":      s.OnFailure,
		})
	}

	restresponse.RestSuccess(c, gin.H{
		"id":          a.ID,
		"name":        a.Name,
		"dialect":     a.Dialect,
		"description": a.Description,
		"notes":       a.Notes,
		"enabled":     a.Enabled,
		"suspended":   a.Suspended,
		"createdAt":   a.CreatedAt,
		"updatedAt":   a.UpdatedAt,
		"steps":       mapped,
	})
}
