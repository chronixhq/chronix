package actionsapi

import (
	"chronix/cxrestapi/apiutil"
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"errors"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func createAction(c *gin.Context) {
	var p actionPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if p.Name == "" || p.Dialect == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "name and dialect are required")
		return
	}

	now := time.Now().UTC()
	a := &models.Action{
		Name:        p.Name,
		Dialect:     p.Dialect,
		Description: p.Description,
		Notes:       p.Notes,
		ActionType:  "database",
		Enabled:     true,
		Suspended:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := db.Q.Transaction(func(tx *db.Query) error {
		if err := tx.Action.Create(a); err != nil {
			return err
		}
		for _, s := range p.Steps {
			st := &models.ActionStep{
				ActionID:       *a.ID,
				StepOrder:      s.Order,
				Name:           s.Name,
				SqlText:        s.SQLText,
				TimeoutSeconds: s.TimeoutSeconds,
				OnFailure:      s.OnFailure,
			}
			if s.Expectation != nil {
				m := datatypes.JSONMap(s.Expectation)
				st.Expectation = &m
			}
			if s.OutputCapture != nil {
				m := datatypes.JSONMap(s.OutputCapture)
				st.OutputCapture = &m
			}
			if err := tx.ActionStep.Create(st); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to create action", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created action", a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"id": a.ID})
}

func updateAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	var p actionPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	a, err := db.Action.Where(db.Action.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}

	a.Name = p.Name
	a.Dialect = p.Dialect
	a.Description = p.Description
	a.Notes = p.Notes
	if p.Enabled != nil {
		if *p.Enabled && !a.Enabled {
		}
		a.Enabled = *p.Enabled
	}
	if p.Suspended != nil {
		a.Suspended = *p.Suspended
	}
	if strings.TrimSpace(a.ActionType) == "" {
		a.ActionType = "database"
	}
	a.UpdatedAt = time.Now().UTC()

	err = db.Q.Transaction(func(tx *db.Query) error {
		if err := tx.Action.Save(a); err != nil {
			return err
		}
		if _, err := tx.ActionStep.Where(tx.ActionStep.ActionID.Eq(id)).Delete(); err != nil {
			return err
		}
		for _, s := range p.Steps {
			st := &models.ActionStep{
				ActionID:       id,
				StepOrder:      s.Order,
				Name:           s.Name,
				SqlText:        s.SQLText,
				TimeoutSeconds: s.TimeoutSeconds,
				OnFailure:      s.OnFailure,
			}
			if s.Expectation != nil {
				m := datatypes.JSONMap(s.Expectation)
				st.Expectation = &m
			}
			if s.OutputCapture != nil {
				m := datatypes.JSONMap(s.OutputCapture)
				st.OutputCapture = &m
			}
			if err := tx.ActionStep.Create(st); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to update action", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated action", a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func patchAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	var p struct {
		Enabled   *bool `json:"enabled"`
		Suspended *bool `json:"suspended"`
	}
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	a, err := db.Action.Where(db.Action.ID.Eq(id)).First()
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
	msg := "Updated action"
	if p.Enabled != nil {
		if *p.Enabled {
			msg = "Enabled action"
		} else {
			msg = "Disabled action"
		}
	}
	_ = activitypkg.RecordUserActivity(user.ID, msg, a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	item, err := db.Action.Where(db.Action.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}

	usageCount, err := db.Job.Where(
		db.Job.ActionID.Eq(id),
		db.Job.Enabled.Is(true),
	).Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to check action usage", err.Error())
		return
	}
	if usageCount > 0 {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot delete action: it is referenced by enabled scheduled jobs")
		return
	}

	if err := db.Q.Transaction(func(tx *db.Query) error {
		if _, err := tx.ActionStep.Where(tx.ActionStep.ActionID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err := tx.Action.Where(tx.Action.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "foreign key") {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Cannot delete action: it is referenced by scheduled jobs", err.Error())
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete action", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted action", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}
