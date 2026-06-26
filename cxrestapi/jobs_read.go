package cxrestapi

import (
	"chronix/internal/db"
	"chronix/internal/scheduler"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func slogWarn(op string, err error, kv ...any) {
	args := append([]any{"component", "cxrestapi", "op", op, "error", err}, kv...)
	slog.Warn("schedule processing", args...)
}

func listJobs(c *gin.Context) {
	type row struct {
		ID            *int64     `json:"id"`
		Name          string     `json:"name"`
		Description   *string    `json:"description"`
		Notes         *string    `json:"notes"`
		ConnectionID  *int64     `json:"connection_id"`
		ActionID      int64      `json:"action_id"`
		TargetKind    string     `json:"target_kind"`
		ShellConnID   *int64     `json:"shell_connection_id"`
		WebtaskConnID *int64     `json:"webtask_connection_id"`
		Enabled       *bool      `json:"enabled"`
		Suspended     *bool      `json:"suspended"`
		CreatedAt     *time.Time `json:"created_at"`
		UpdatedAt     *time.Time `json:"updated_at"`
		ScheduleJSON  *string    `json:"schedule_json"`
		AlertEmails   *string    `json:"alert_emails"`
		AlertPhones   *string    `json:"alert_phones"`
		OnSuccess     *bool      `json:"notify_on_success"`
		OnError       *bool      `json:"notify_on_error"`
		IncludeOutput *bool      `json:"notify_include_output"`
	}
	var rows []row
	if err := db.Job.
		Select(
			db.Job.ID,
			db.Job.Name,
			db.Job.Description,
			db.Job.Notes,
			db.Job.ConnectionID,
			db.Job.ActionID,
			db.Job.TargetKind,
			db.Job.ShellConnectionID,
			db.Job.WebtaskConnectionID,
			db.Job.Enabled,
			db.Job.Suspended,
			db.Job.CreatedAt,
			db.Job.UpdatedAt,
			db.Job.ScheduleJSON,
			db.Job.AlertEmails,
			db.Job.AlertPhones,
			db.Job.NotifyOnSuccess,
			db.Job.NotifyOnError,
			db.Job.NotifyIncludeOutput,
		).
		Order(db.Job.CreatedAt.Desc()).
		Scan(&rows); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list jobs", err.Error())
		return
	}

	lastByJob := map[int64]*time.Time{}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		if r.ID != nil {
			ids = append(ids, *r.ID)
		}
	}
	if len(ids) > 0 {
		type lastRow struct {
			JobID      int64      `json:"job_id"`
			QueuedAt   *time.Time `json:"queued_at"`
			StartedAt  *time.Time `json:"started_at"`
			FinishedAt *time.Time `json:"finished_at"`
		}
		var lr []lastRow
		if err := db.JobRun.
			Select(db.JobRun.JobID, db.JobRun.QueuedAt, db.JobRun.StartedAt, db.JobRun.FinishedAt).
			Where(db.JobRun.JobID.In(ids...)).
			Order(db.JobRun.StartedAt.Desc(), db.JobRun.QueuedAt.Desc(), db.JobRun.FinishedAt.Desc()).
			Scan(&lr); err == nil {
			for _, x := range lr {
				if _, ok := lastByJob[x.JobID]; ok {
					continue
				}
				if x.FinishedAt != nil {
					lastByJob[x.JobID] = x.FinishedAt
				} else if x.StartedAt != nil {
					lastByJob[x.JobID] = x.StartedAt
				} else if x.QueuedAt != nil {
					lastByJob[x.JobID] = x.QueuedAt
				}
			}
		}
	}

	resp := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		schedule, nextAt := scheduleAndNextRun(r.ScheduleJSON, r.Enabled, lastByJob[derefInt64(r.ID)])
		resp = append(resp, gin.H{
			"id":                    r.ID,
			"name":                  r.Name,
			"description":           r.Description,
			"notes":                 r.Notes,
			"connectionId":          r.ConnectionID,
			"actionId":              r.ActionID,
			"targetKind":            r.TargetKind,
			"shellConnectionId":     r.ShellConnID,
			"webtaskConnectionId":   r.WebtaskConnID,
			"schedule":              schedule,
			"enabled":               r.Enabled,
			"suspended":             r.Suspended,
			"nextRunAt":             nextAt,
			"lastRunAt":             lastByJob[derefInt64(r.ID)],
			"createdAt":             r.CreatedAt,
			"updatedAt":             r.UpdatedAt,
			"alert_emails":          r.AlertEmails,
			"alert_phones":          r.AlertPhones,
			"notify_on_success":     r.OnSuccess,
			"notify_on_error":       r.OnError,
			"notify_include_output": r.IncludeOutput,
		})
	}
	restresponse.RestSuccess(c, resp)
}

func getJob(c *gin.Context) {
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
	vars, _ := db.JobVariable.Where(db.JobVariable.JobID.Eq(*item.ID)).Order(db.JobVariable.Name.Asc()).Find()
	varsResp := make([]gin.H, 0, len(vars))
	for _, v := range vars {
		varsResp = append(varsResp, gin.H{"name": v.Name, "value": v.Value})
	}

	var schedRaw *string
	_ = db.Job.Select(db.Job.ScheduleJSON).Where(db.Job.ID.Eq(id)).Scan(&schedRaw)
	lastAt := lastRunForJob(*item.ID)
	schedule, nextAt := scheduleAndNextRun(schedRaw, item.Enabled, lastAt)

	resp := gin.H{
		"id":                    item.ID,
		"name":                  item.Name,
		"description":           item.Description,
		"notes":                 item.Notes,
		"connectionId":          item.ConnectionID,
		"actionId":              item.ActionID,
		"targetKind":            item.TargetKind,
		"shellConnectionId":     item.ShellConnectionID,
		"webtaskConnectionId":   item.WebtaskConnectionID,
		"schedule":              schedule,
		"enabled":               item.Enabled,
		"suspended":             item.Suspended,
		"variables":             varsResp,
		"nextRunAt":             nextAt,
		"lastRunAt":             lastAt,
		"createdAt":             item.CreatedAt,
		"updatedAt":             item.UpdatedAt,
		"alert_emails":          item.AlertEmails,
		"alert_phones":          item.AlertPhones,
		"notify_on_success":     item.NotifyOnSuccess,
		"notify_on_error":       item.NotifyOnError,
		"notify_include_output": item.NotifyIncludeOutput,
	}
	restresponse.RestSuccess(c, resp)
}

func listJobRuns(c *gin.Context) {
	id := atoi64(c.Param("id"))
	if _, err := db.Job.Where(db.Job.ID.Eq(id)).First(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Job not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load job", err.Error())
		return
	}
	limit := atoi64(c.DefaultQuery("limit", "20"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}
	offset := atoi64(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}
	status := c.DefaultQuery("status", "")
	q := db.JobRun.Where(db.JobRun.JobID.Eq(id))
	if status != "" {
		q = q.Where(db.JobRun.Status.Eq(status))
	}
	type row struct {
		RunUID      string     `json:"run_uid"`
		Status      string     `json:"status"`
		QueuedAt    time.Time  `json:"queued_at"`
		StartedAt   *time.Time `json:"started_at"`
		FinishedAt  *time.Time `json:"finished_at"`
		Message     *string    `json:"message"`
		TriggeredBy *int64     `json:"triggered_by"`
	}
	total, err := q.Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to count job runs", err)
		return
	}
	var rows []row
	if err := q.Select(
		db.JobRun.RunUID,
		db.JobRun.Status,
		db.JobRun.QueuedAt,
		db.JobRun.StartedAt,
		db.JobRun.FinishedAt,
		db.JobRun.Message,
		db.JobRun.TriggeredBy,
	).Order(db.JobRun.QueuedAt.Desc()).Limit(int(limit)).Offset(int(offset)).Scan(&rows); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list job runs", err)
		return
	}
	resp := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, gin.H{
			"runId":       r.RunUID,
			"status":      r.Status,
			"queuedAt":    r.QueuedAt,
			"startedAt":   r.StartedAt,
			"finishedAt":  r.FinishedAt,
			"message":     r.Message,
			"triggeredBy": r.TriggeredBy,
		})
	}
	restresponse.RestSuccess(c, gin.H{"items": resp, "total": total})
}

func scheduleAndNextRun(scheduleJSON *string, enabled *bool, lastAt *time.Time) (any, *time.Time) {
	var schedule any
	var nextAt *time.Time
	if scheduleJSON == nil || *scheduleJSON == "" {
		return schedule, nextAt
	}
	var js any
	if err := json.Unmarshal([]byte(*scheduleJSON), &js); err == nil {
		schedule = js
	}
	if enabled != nil && *enabled {
		if t, err := scheduler.NextRunTime([]byte(*scheduleJSON)); err == nil && !t.IsZero() {
			tCopy := t
			nextAt = &tCopy
		}
	}
	if nextAt != nil && lastAt != nil && nextAt.Truncate(time.Minute).Equal(lastAt.Truncate(time.Minute)) {
		if t2, err := scheduler.NextRunTime([]byte(*scheduleJSON), lastAt.Add(time.Minute)); err == nil && !t2.IsZero() {
			tCopy := t2
			nextAt = &tCopy
		} else {
			nextAt = nil
		}
	}
	return schedule, nextAt
}

func lastRunForJob(jobID int64) *time.Time {
	type lastRow struct {
		QueuedAt   *time.Time
		StartedAt  *time.Time
		FinishedAt *time.Time
	}
	var lr lastRow
	if err := db.JobRun.
		Select(db.JobRun.QueuedAt, db.JobRun.StartedAt, db.JobRun.FinishedAt).
		Where(db.JobRun.JobID.Eq(jobID)).
		Order(db.JobRun.StartedAt.Desc(), db.JobRun.QueuedAt.Desc(), db.JobRun.FinishedAt.Desc()).
		Limit(1).
		Scan(&lr); err == nil {
		if lr.FinishedAt != nil {
			return lr.FinishedAt
		}
		if lr.StartedAt != nil {
			return lr.StartedAt
		}
		if lr.QueuedAt != nil {
			return lr.QueuedAt
		}
	}
	return nil
}

func derefInt64(v *int64) int64 {
	if v == nil {
		return 0
	}
	return *v
}
