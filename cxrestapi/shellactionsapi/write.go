package shellactionsapi

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

func createShellAction(c *gin.Context) {
	var p shellActionPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if strings.TrimSpace(p.Name) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "name is required")
		return
	}

	now := time.Now().UTC()
	a := &models.Action{
		Name:        p.Name,
		Description: p.Description,
		Notes:       p.Notes,
		ActionType:  "shell",
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
			capBytes := capOutputCaptureBytes(s.OutputCaptureMaxBytes)
			st := &models.ShellActionStep{
				ActionID:              *a.ID,
				StepOrder:             &s.Order,
				Name:                  s.Name,
				RunMode:               &s.RunMode,
				Command:               s.Command,
				ScriptText:            s.ScriptText,
				ShellPath:             &s.ShellPath,
				WorkingDir:            s.WorkingDir,
				TimeoutSeconds:        s.TimeoutSeconds,
				OutputCaptureMaxBytes: &capBytes,
				OutputTruncation:      &s.OutputTruncation,
				OnFailure:             s.OnFailure,
			}
			if len(s.Expectation) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.Expectation {
					m[k] = v
				}
				st.Expectation = &m
			}
			if len(s.OutputCapture) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.OutputCapture {
					m[k] = v
				}
				st.OutputCapture = &m
			}
			if len(s.Env) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.Env {
					m[k] = v
				}
				st.EnvJSON = &m
			}
			if err := tx.ShellActionStep.Create(st); err != nil {
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
	_ = activitypkg.RecordUserActivity(user.ID, "Created shell action", a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"id": a.ID})
}

func updateShellAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	var p shellActionPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	a, err := db.Action.Where(db.Action.ID.Eq(id), db.Action.ActionType.Eq("shell")).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}
	a.Name = p.Name
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
	a.UpdatedAt = time.Now().UTC()
	err = db.Q.Transaction(func(tx *db.Query) error {
		if err := tx.Action.Save(a); err != nil {
			return err
		}
		if _, err := tx.ShellActionStep.Where(tx.ShellActionStep.ActionID.Eq(id)).Delete(); err != nil {
			return err
		}
		for _, s := range p.Steps {
			capBytes := capOutputCaptureBytes(s.OutputCaptureMaxBytes)
			st := &models.ShellActionStep{
				ActionID:              id,
				StepOrder:             &s.Order,
				Name:                  s.Name,
				RunMode:               &s.RunMode,
				Command:               s.Command,
				ScriptText:            s.ScriptText,
				ShellPath:             &s.ShellPath,
				WorkingDir:            s.WorkingDir,
				TimeoutSeconds:        s.TimeoutSeconds,
				OutputCaptureMaxBytes: &capBytes,
				OutputTruncation:      &s.OutputTruncation,
				OnFailure:             s.OnFailure,
			}
			if len(s.Expectation) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.Expectation {
					m[k] = v
				}
				st.Expectation = &m
			}
			if len(s.OutputCapture) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.OutputCapture {
					m[k] = v
				}
				st.OutputCapture = &m
			}
			if len(s.Env) > 0 {
				m := datatypes.JSONMap{}
				for k, v := range s.Env {
					m[k] = v
				}
				st.EnvJSON = &m
			}
			if err := tx.ShellActionStep.Create(st); err != nil {
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
	_ = activitypkg.RecordUserActivity(user.ID, "Updated shell action", a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func patchShellAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	var p struct {
		Enabled   *bool `json:"enabled"`
		Suspended *bool `json:"suspended"`
	}
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	a, err := db.Action.Where(db.Action.ID.Eq(id), db.Action.ActionType.Eq("shell")).First()
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
	msg := "Updated shell action"
	if p.Enabled != nil {
		if *p.Enabled {
			msg = "Enabled shell action"
		} else {
			msg = "Disabled shell action"
		}
	}
	_ = activitypkg.RecordUserActivity(user.ID, msg, a.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteShellAction(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	item, err := db.Action.Where(db.Action.ID.Eq(id), db.Action.ActionType.Eq("shell")).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Action not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load action", err.Error())
		return
	}
	usageCount, err := db.Job.Where(db.Job.ActionID.Eq(id), db.Job.Enabled.Is(true)).Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to check usage", err.Error())
		return
	}
	if usageCount > 0 {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot delete action: it is referenced by enabled jobs")
		return
	}
	if err := db.Q.Transaction(func(tx *db.Query) error {
		if _, err := tx.ShellActionStep.Where(tx.ShellActionStep.ActionID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err := tx.Action.Where(tx.Action.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete action", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted shell action", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}
