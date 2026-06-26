package webtaskactionsapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/execution"
	"chronix/internal/secret"
	"errors"
	"strings"
	"time"

	"chronix/cxrestapi/apiutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func listWebtaskActions(c *gin.Context) {
	rows, err := db.Action.Where(db.Action.ActionType.Eq("webtask")).Order(db.Action.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to list actions", err.Error())
		return
	}

	ids := make([]int64, 0, len(rows))
	for _, a := range rows {
		if a.ID != nil {
			ids = append(ids, *a.ID)
		}
	}

	stepsByAction := map[int64][]gin.H{}
	if len(ids) > 0 {
		steps, err := db.WebtaskActionStep.Where(db.WebtaskActionStep.ActionID.In(ids...)).Order(db.WebtaskActionStep.StepOrder.Asc()).Find()
		if err == nil {
			for _, s := range steps {
				stepsByAction[s.ActionID] = append(stepsByAction[s.ActionID], gin.H{
					"id":              s.ID,
					"order":           s.StepOrder,
					"name":            s.Name,
					"method":          s.Method,
					"url":             s.URL,
					"headers":         s.Headers,
					"body":            s.Body,
					"timeoutSeconds":  s.TimeoutSeconds,
					"expectation":     s.Expectation,
					"responseCapture": s.ResponseCapture,
					"onFailure":       s.OnFailure,
				})
			}
		}
	}

	resp := make([]gin.H, 0, len(rows))
	for _, row := range rows {
		steps := stepsByAction[pickID(row.ID)]
		resp = append(resp, gin.H{
			"id":          row.ID,
			"name":        row.Name,
			"description": row.Description,
			"notes":       row.Notes,
			"actionType":  row.ActionType,
			"createdAt":   row.CreatedAt,
			"updatedAt":   row.UpdatedAt,
			"steps":       steps,
		})
	}
	restresponse.RestSuccess(c, resp)
}

func getWebtaskAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	row, err := db.Action.Where(db.Action.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "action not found")
		return
	}
	steps, _ := db.WebtaskActionStep.Where(db.WebtaskActionStep.ActionID.Eq(id)).Order(db.WebtaskActionStep.StepOrder.Asc()).Find()

	mappedSteps := make([]gin.H, 0, len(steps))
	for _, s := range steps {
		mappedSteps = append(mappedSteps, gin.H{
			"id":              s.ID,
			"order":           s.StepOrder,
			"name":            s.Name,
			"method":          s.Method,
			"url":             s.URL,
			"headers":         s.Headers,
			"body":            s.Body,
			"timeoutSeconds":  s.TimeoutSeconds,
			"expectation":     s.Expectation,
			"responseCapture": s.ResponseCapture,
			"onFailure":       s.OnFailure,
		})
	}

	restresponse.RestSuccess(c, gin.H{
		"id":          row.ID,
		"name":        row.Name,
		"description": row.Description,
		"notes":       row.Notes,
		"actionType":  row.ActionType,
		"createdAt":   row.CreatedAt,
		"updatedAt":   row.UpdatedAt,
		"steps":       mappedSteps,
	})
}

func createWebtaskAction(c *gin.Context) {
	var p webtaskActionPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}

	row := &models.Action{
		Name:        p.Name,
		Description: utilities.Ptr(p.Description),
		Notes:       utilities.Ptr(p.Notes),
		ActionType:  "webtask",
		Dialect:     "generic",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := db.Action.Create(row); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to create action", err.Error())
		return
	}

	_, _ = db.WebtaskActionStep.Where(db.WebtaskActionStep.ActionID.Eq(*row.ID)).Delete()
	for _, s := range p.Steps {
		step := &models.WebtaskActionStep{
			ActionID:        *row.ID,
			StepOrder:       &s.StepOrder,
			Name:            s.Name,
			Method:          s.Method,
			URL:             s.URL,
			Headers:         s.Headers,
			Body:            s.Body,
			TimeoutSeconds:  s.TimeoutSeconds,
			Expectation:     s.Expectation,
			ResponseCapture: s.ResponseCapture,
			OnFailure:       s.OnFailure,
		}
		_ = db.WebtaskActionStep.Create(step)
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created webtask action", row.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, row)
}

func updateWebtaskAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	old, err := db.Action.Where(db.Action.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "action not found")
		return
	}

	var p webtaskActionPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}

	old.Name = p.Name
	old.Description = utilities.Ptr(p.Description)
	old.Notes = utilities.Ptr(p.Notes)
	old.UpdatedAt = time.Now()

	if _, err := db.Action.Where(db.Action.ID.Eq(id)).Updates(old); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to update action", err.Error())
		return
	}

	_, _ = db.WebtaskActionStep.Where(db.WebtaskActionStep.ActionID.Eq(id)).Delete()
	for _, s := range p.Steps {
		step := &models.WebtaskActionStep{
			ActionID:        id,
			StepOrder:       &s.StepOrder,
			Name:            s.Name,
			Method:          s.Method,
			URL:             s.URL,
			Headers:         s.Headers,
			Body:            s.Body,
			TimeoutSeconds:  s.TimeoutSeconds,
			Expectation:     s.Expectation,
			ResponseCapture: s.ResponseCapture,
			OnFailure:       s.OnFailure,
		}
		_ = db.WebtaskActionStep.Create(step)
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated webtask action", old.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, old)
}

func patchWebtaskAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	var p struct {
		Enabled   *bool `json:"enabled"`
		Suspended *bool `json:"suspended"`
	}
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	a, err := db.Action.Where(db.Action.ID.Eq(id), db.Action.ActionType.Eq("webtask")).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}

	if p.Enabled != nil {
		if *p.Enabled && !a.Enabled {
		}
		a.Enabled = *p.Enabled
	}
	if p.Suspended != nil {
		a.Suspended = *p.Suspended
	}

	a.UpdatedAt = time.Now().UTC()
	if err := db.Action.Save(a); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to patch action", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	msg := "Updated webtask action"
	if p.Enabled != nil {
		if *p.Enabled {
			msg = "Enabled webtask action"
		} else {
			msg = "Disabled webtask action"
		}
	}
	_ = activitypkg.RecordUserActivity(user.ID, msg, a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteWebtaskAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	item, err := db.Action.Where(db.Action.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "action not found")
		return
	}
	if err := db.Q.Transaction(func(tx *db.Query) error {
		if _, err := tx.WebtaskActionStep.Where(tx.WebtaskActionStep.ActionID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err := tx.Action.Where(tx.Action.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to delete action", err.Error())
		return
	}
	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted webtask action", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func testWebtaskAction(c *gin.Context) {
	var p webtaskActionTestPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}

	conn, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(p.ConnectionID)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "connection not found")
		return
	}

	if conn.AuthConfig != nil {
		auth := *conn.AuthConfig
		authType := strings.ToLower(conn.AuthType)
		switch authType {
		case "basic":
			if pass, ok := auth["password"].(string); ok && pass != "" {
				auth["password"], _ = secret.Decrypt(pass)
			}
		case "bearer":
			if token, ok := auth["token"].(string); ok && token != "" {
				auth["token"], _ = secret.Decrypt(token)
			}
		case "header":
			if val, ok := auth["header_value"].(string); ok && val != "" {
				auth["header_value"], _ = secret.Decrypt(val)
			}
		}
	}

	steps := make([]models.WebtaskActionStep, 0, len(p.Steps))
	for _, step := range p.Steps {
		steps = append(steps, models.WebtaskActionStep(step))
	}

	results, err := execution.TestWebTaskAction(c.Request.Context(), steps, conn, p.Variables)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Test failed", err.Error())
		return
	}

	restresponse.RestSuccess(c, results)
}
