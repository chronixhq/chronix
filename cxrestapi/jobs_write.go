package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	jobrunpkg "chronix/internal/jobrun"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type jobVarPayload struct {
	Name  string  `json:"name"`
	Value *string `json:"value"`
}

type jobPayload struct {
	Name                string          `json:"name"`
	Description         *string         `json:"description"`
	Notes               *string         `json:"notes"`
	ConnectionID        int64           `json:"connection_id"`
	ActionID            int64           `json:"action_id"`
	Schedule            json.RawMessage `json:"schedule"`
	Enabled             *bool           `json:"enabled"`
	Suspended           *bool           `json:"suspended"`
	Variables           []jobVarPayload `json:"variables"`
	TargetKind          string          `json:"target_kind"`
	ShellConnectionID   int64           `json:"shell_connection_id"`
	WebtaskConnectionID int64           `json:"webtask_connection_id"`

	AlertEmails         *string `json:"alert_emails"`
	AlertPhones         *string `json:"alert_phones"`
	NotifyOnSuccess     *bool   `json:"notify_on_success"`
	NotifyOnError       *bool   `json:"notify_on_error"`
	NotifyIncludeOutput *bool   `json:"notify_include_output"`
}

func createJob(c *gin.Context) {
	var p jobPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	tk := strings.TrimSpace(strings.ToLower(p.TargetKind))
	if tk == "" {
		tk = "database"
	}
	if tk != "database" && tk != "shell" && tk != "webtask" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "target_kind must be 'database', 'shell', or 'webtask'")
		return
	}
	if p.Name == "" || p.ActionID == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "name and action_id are required")
		return
	}
	if tk == "database" && p.ConnectionID == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "connection_id is required for database jobs")
		return
	}
	if tk == "shell" && p.ShellConnectionID == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "shell_connection_id is required for shell jobs")
		return
	}
	if tk == "webtask" && p.WebtaskConnectionID == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "webtask_connection_id is required for webtask jobs")
		return
	}
	if act, err := db.Action.Where(db.Action.ID.Eq(p.ActionID)).First(); err == nil {
		if (tk == "database" && strings.ToLower(act.ActionType) != "database") || (tk == "shell" && strings.ToLower(act.ActionType) != "shell") || (tk == "webtask" && strings.ToLower(act.ActionType) != "webtask") {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "action type does not match target_kind")
			return
		}
	}
	if len(p.Schedule) == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "schedule is required")
		return
	}

	now := time.Now().UTC()
	item := newJobModelFromPayload(&p, tk, now)
	if err := db.Job.Create(item); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to create job", err.Error())
		return
	}
	if _, err := db.Job.Where(db.Job.ID.Eq(*item.ID)).Update(db.Job.ScheduleJSON, string(p.Schedule)); err != nil {
		slogWarn("persist schedule_json", err)
	}
	if len(p.Variables) > 0 {
		for _, v := range p.Variables {
			name := strings.TrimSpace(v.Name)
			if name == "" {
				continue
			}
			_ = db.JobVariable.Create(&models.JobVariable{JobID: *item.ID, Name: name, Value: v.Value})
		}
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created job", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"id": item.ID})
}

func updateJob(c *gin.Context) {
	id := atoi64(c.Param("id"))
	var p jobPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	item, err := db.Job.Where(db.Job.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load job", err.Error())
		return
	}
	item.Name = p.Name
	item.Description = p.Description
	item.Notes = p.Notes
	item.ActionID = p.ActionID
	tk := strings.TrimSpace(strings.ToLower(p.TargetKind))
	if tk == "" {
		tk = item.TargetKind
		if tk == "" {
			tk = "database"
		}
	}
	if tk != "database" && tk != "shell" && tk != "webtask" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "target_kind must be 'database', 'shell', or 'webtask'")
		return
	}
	item.TargetKind = tk
	if tk == "database" {
		cid := p.ConnectionID
		item.ConnectionID = &cid
		item.ShellConnectionID = nil
		item.WebtaskConnectionID = nil
	} else if tk == "shell" {
		if p.ShellConnectionID == 0 {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "shell_connection_id is required for shell jobs")
			return
		}
		item.ConnectionID = nil
		scid := p.ShellConnectionID
		item.ShellConnectionID = &scid
		item.WebtaskConnectionID = nil
	} else {
		if p.WebtaskConnectionID == 0 {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "webtask_connection_id is required for webtask jobs")
			return
		}
		item.ConnectionID = nil
		item.ShellConnectionID = nil
		wcid := p.WebtaskConnectionID
		item.WebtaskConnectionID = &wcid
	}
	if p.Enabled != nil {
		if *p.Enabled && (item.Enabled == nil || !*item.Enabled) {
		}
		item.Enabled = p.Enabled
	}
	if p.Suspended != nil {
		item.Suspended = *p.Suspended
	}
	item.AlertEmails = p.AlertEmails
	item.AlertPhones = p.AlertPhones
	item.NotifyOnSuccess = p.NotifyOnSuccess
	item.NotifyOnError = p.NotifyOnError
	item.NotifyIncludeOutput = p.NotifyIncludeOutput
	item.UpdatedAt = time.Now().UTC()

	err = db.Q.Transaction(func(tx *db.Query) error {
		if err := tx.Job.Save(item); err != nil {
			return err
		}
		if len(p.Schedule) > 0 {
			if _, err := tx.Job.Where(tx.Job.ID.Eq(id)).Update(tx.Job.ScheduleJSON, string(p.Schedule)); err != nil {
				return err
			}
		}
		if _, err := tx.JobVariable.Where(tx.JobVariable.JobID.Eq(id)).Delete(); err != nil {
			return err
		}
		for _, v := range p.Variables {
			name := strings.TrimSpace(v.Name)
			if name == "" {
				continue
			}
			if err := tx.JobVariable.Create(&models.JobVariable{JobID: id, Name: name, Value: v.Value}); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to update job", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated job", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteJob(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.Job.Where(db.Job.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load job", err.Error())
		return
	}
	err = db.Q.Transaction(func(tx *db.Query) error {
		if _, err := tx.JobVariable.Where(tx.JobVariable.JobID.Eq(id)).Delete(); err != nil {
			return err
		}
		if _, err := tx.Job.Where(tx.Job.ID.Eq(id)).Delete(); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete job", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted job", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func enableJob(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.Job.Where(db.Job.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
		return
	}
	if _, err = db.Job.Where(db.Job.ID.Eq(id)).Update(db.Job.Enabled, true); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to enable job", err)
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Enabled job", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func disableJob(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.Job.Where(db.Job.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
		return
	}
	if _, err = db.Job.Where(db.Job.ID.Eq(id)).Update(db.Job.Enabled, false); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to disable job", err)
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Disabled job", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func runNowJob(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.Job.Where(db.Job.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load job", err.Error())
		return
	}
	user := userFromGinContext(c)
	runID, err := jobrunpkg.EnqueueJobRun(id, user.ID)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to queue job run", err.Error())
		return
	}
	_ = activitypkg.RecordUserActivity(user.ID, "Manual job run", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"status": "queued", "runId": runID})
}

func newJobModelFromPayload(p *jobPayload, targetKind string, now time.Time) *models.Job {
	item := &models.Job{
		Name:                p.Name,
		Description:         p.Description,
		Notes:               p.Notes,
		ActionID:            p.ActionID,
		TargetKind:          targetKind,
		ConnectionID:        nil,
		ShellConnectionID:   nil,
		WebtaskConnectionID: nil,
		Enabled:             utilities.Ptr(pickBool(p.Enabled, true)),
		Suspended:           false,
		AlertEmails:         p.AlertEmails,
		AlertPhones:         p.AlertPhones,
		NotifyOnSuccess:     p.NotifyOnSuccess,
		NotifyOnError:       p.NotifyOnError,
		NotifyIncludeOutput: p.NotifyIncludeOutput,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
	if targetKind == "database" {
		cid := p.ConnectionID
		item.ConnectionID = &cid
	} else if targetKind == "shell" {
		scid := p.ShellConnectionID
		item.ShellConnectionID = &scid
	} else {
		wcid := p.WebtaskConnectionID
		item.WebtaskConnectionID = &wcid
	}
	return item
}
