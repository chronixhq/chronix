package cxrestapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	progresspkg "chronix/internal/progress"
	"strconv"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func getRunProgress(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if runID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid runId")
		return
	}
	if snap, events, ok := progresspkg.GetRunProgress(runID); ok {
		restresponse.RestSuccess(c, gin.H{"snapshot": snap, "events": events})
		return
	}
	jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runID)).First()
	if err != nil || jr == nil || jr.ID == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "run not found")
		return
	}
	var stepRows []struct {
		ID         int64      `json:"id"`
		StepOrder  int64      `json:"step_order"`
		StepName   string     `json:"step_name"`
		Status     string     `json:"status"`
		StartedAt  *time.Time `json:"started_at"`
		FinishedAt *time.Time `json:"finished_at"`
	}
	if err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(*jr.ID)).
		Select(db.JobRunStep.ID, db.JobRunStep.StepOrder, db.JobRunStep.StepName, db.JobRunStep.Status, db.JobRunStep.StartedAt, db.JobRunStep.FinishedAt).
		Order(db.JobRunStep.StepOrder.Asc()).
		Scan(&stepRows); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "load steps failed", err)
		return
	}
	var curIdx *int
	var curName *string
	for _, s := range stepRows {
		if s.Status == "running" {
			idx := int(s.StepOrder)
			name := s.StepName
			curIdx = &idx
			curName = &name
			break
		}
	}
	if curIdx == nil && len(stepRows) > 0 {
		last := stepRows[len(stepRows)-1]
		idx := int(last.StepOrder)
		name := last.StepName
		curIdx = &idx
		curName = &name
	}
	eventsLimit := 200
	if v := c.Query("events_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			eventsLimit = n
		}
	}
	var eventsRaw []struct {
		CreatedAt time.Time
		Kind      string
		Message   *string
		StepID    *int64
	}
	if err := db.JobRunEvent.Where(db.JobRunEvent.RunID.Eq(*jr.ID)).
		Select(db.JobRunEvent.CreatedAt, db.JobRunEvent.Kind, db.JobRunEvent.Message, db.JobRunEvent.StepID).
		Order(db.JobRunEvent.CreatedAt.Asc()).
		Limit(eventsLimit).
		Scan(&eventsRaw); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "load events failed", err)
		return
	}
	stepMeta := map[int64]struct {
		order int
		name  string
	}{}
	for _, s := range stepRows {
		stepMeta[s.ID] = struct {
			order int
			name  string
		}{order: int(s.StepOrder), name: s.StepName}
	}
	type liteEvent struct {
		Ts        time.Time `json:"ts"`
		Type      string    `json:"type"`
		StepIndex *int      `json:"stepIndex,omitempty"`
		StepName  *string   `json:"stepName,omitempty"`
		Message   *string   `json:"message,omitempty"`
	}
	lite := make([]liteEvent, 0, len(eventsRaw))
	for _, ev := range eventsRaw {
		var idxPtr *int
		var namePtr *string
		if ev.StepID != nil {
			if m, ok := stepMeta[*ev.StepID]; ok {
				idx := m.order
				name := m.name
				idxPtr = &idx
				namePtr = &name
			}
		}
		lite = append(lite, liteEvent{Ts: ev.CreatedAt, Type: ev.Kind, StepIndex: idxPtr, StepName: namePtr, Message: ev.Message})
	}
	snapshot := progresspkg.Snapshot{
		RunID:     runID,
		JobID:     *jr.JobID,
		Status:    *jr.Status,
		StartedAt: jr.StartedAt,
		UpdatedAt: func() time.Time {
			if len(eventsRaw) > 0 {
				return eventsRaw[len(eventsRaw)-1].CreatedAt
			}
			if jr.QueuedAt != nil {
				return *jr.QueuedAt
			}
			return time.Now().UTC()
		}(),
		Message:         jr.Message,
		CurrentStep:     curIdx,
		CurrentStepName: curName,
	}
	restresponse.RestSuccess(c, gin.H{"snapshot": snapshot, "events": lite})
}

func listRuns(c *gin.Context) {
	q := db.JobRun.Where(db.JobRun.ID.IsNotNull())
	if v := strings.TrimSpace(c.Query("job_id")); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			q = q.Where(db.JobRun.JobID.Eq(id))
		}
	}
	if v := strings.TrimSpace(c.Query("status")); v != "" {
		q = q.Where(db.JobRun.Status.Eq(v))
	}
	if v := strings.TrimSpace(c.Query("since")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q = q.Where(db.JobRun.QueuedAt.Gte(t))
		}
	}
	if v := strings.TrimSpace(c.Query("until")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q = q.Where(db.JobRun.QueuedAt.Lte(t))
		}
	}
	if v := strings.TrimSpace(c.Query("started_from")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q = q.Where(db.JobRun.StartedAt.Gte(t))
		}
	}
	if v := strings.TrimSpace(c.Query("started_to")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q = q.Where(db.JobRun.StartedAt.Lte(t))
		}
	}
	if v := strings.TrimSpace(c.Query("q")); v != "" {
		like := "%" + v + "%"
		q = q.Where(db.JobRun.RunUID.Like(like)).Or(db.JobRun.JobName.Like(like)).Or(db.JobRun.Status.Like(like)).Or(db.JobRun.Message.Like(like))
	}
	total, err := q.Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "count runs failed", err)
		return
	}
	limit := 50
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	offset := 0
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	type row struct {
		RunUID        string     `json:"run_uid"`
		JobID         int64      `json:"job_id"`
		JobName       *string    `json:"job_name"`
		Status        string     `json:"status"`
		QueuedAt      time.Time  `json:"queued_at"`
		StartedAt     *time.Time `json:"started_at"`
		FinishedAt    *time.Time `json:"finished_at"`
		TriggeredBy   *int64     `json:"triggered_by"`
		TriggerSource *string    `json:"trigger_source"`
		Message       *string    `json:"message"`
		RowsAffected  *int64     `json:"rows_affected"`
	}
	var rows []row
	err = q.Select(
		db.JobRun.RunUID, db.JobRun.JobID, db.JobRun.JobName, db.JobRun.Status, db.JobRun.QueuedAt,
		db.JobRun.StartedAt, db.JobRun.FinishedAt, db.JobRun.TriggeredBy, db.JobRun.TriggerSource,
		db.JobRun.Message, db.JobRun.RowsAffected,
	).Order(db.JobRun.QueuedAt.Desc()).Limit(limit).Offset(offset).Scan(&rows)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "list runs failed", err)
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		var durMs *int64
		if r.StartedAt != nil && r.FinishedAt != nil {
			d := r.FinishedAt.Sub(*r.StartedAt).Milliseconds()
			durMs = &d
		}
		items = append(items, gin.H{
			"runId":         r.RunUID,
			"jobId":         r.JobID,
			"jobName":       r.JobName,
			"status":        r.Status,
			"queuedAt":      r.QueuedAt,
			"startedAt":     r.StartedAt,
			"finishedAt":    r.FinishedAt,
			"durationMs":    durMs,
			"triggeredBy":   r.TriggeredBy,
			"triggerSource": r.TriggerSource,
			"message":       r.Message,
			"rowsAffected":  r.RowsAffected,
		})
	}
	restresponse.RestSuccess(c, gin.H{"items": items, "total": total})
}

func getRunDetail(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if runID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid runId")
		return
	}
	jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runID)).First()
	if err != nil || jr == nil || jr.ID == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "run not found")
		return
	}
	stepsResp, err := loadRunSteps(*jr.ID)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "load steps failed", err)
		return
	}
	eventsLimit := 500
	if v := c.Query("events_limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 2000 {
			eventsLimit = n
		}
	}
	eventsOffset := 0
	if v := c.Query("events_offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			eventsOffset = n
		}
	}
	qe := db.JobRunEvent.Where(db.JobRunEvent.RunID.Eq(*jr.ID))
	if v := strings.TrimSpace(c.Query("events_since")); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			qe = qe.Where(db.JobRunEvent.CreatedAt.Gte(t))
		}
	}
	var events []struct {
		CreatedAt time.Time          `json:"createdAt"`
		Kind      string             `json:"kind"`
		Message   *string            `json:"message"`
		StepID    *int64             `json:"stepId"`
		Data      *datatypes.JSONMap `json:"data"`
	}
	err = qe.Select(db.JobRunEvent.CreatedAt, db.JobRunEvent.Kind, db.JobRunEvent.Message, db.JobRunEvent.StepID, db.JobRunEvent.Data).
		Order(db.JobRunEvent.CreatedAt.Asc()).Limit(eventsLimit).Offset(eventsOffset).Scan(&events)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "load events failed", err)
		return
	}
	runObj := gin.H{
		"runId":         *jr.RunUID,
		"jobId":         *jr.JobID,
		"jobName":       jr.JobName,
		"status":        *jr.Status,
		"queuedAt":      *jr.QueuedAt,
		"startedAt":     jr.StartedAt,
		"finishedAt":    jr.FinishedAt,
		"triggeredBy":   jr.TriggeredBy,
		"triggerSource": jr.TriggerSource,
		"message":       jr.Message,
		"rowsAffected":  jr.RowsAffected,
		"summary":       jr.Summary,
	}
	enrichRunObject(runObj, jr)
	restresponse.RestSuccess(c, gin.H{"run": runObj, "steps": stepsResp, "events": events})
}

func loadRunSteps(runDBID int64) ([]gin.H, error) {
	var steps []struct {
		ID             int64              `json:"id"`
		StepOrder      int64              `json:"stepOrder"`
		StepName       string             `json:"stepName"`
		Status         string             `json:"status"`
		StartedAt      *time.Time         `json:"startedAt"`
		FinishedAt     *time.Time         `json:"finishedAt"`
		SQLText        *string            `json:"sqlText"`
		TimeoutSeconds *int64             `json:"timeoutSeconds"`
		RowsCount      *int64             `json:"rowsCount"`
		RowsAffected   *int64             `json:"rowsAffected"`
		Expectation    *datatypes.JSONMap `json:"expectation"`
		ExpectOK       *bool              `json:"expectOk"`
		ExpectMessage  *string            `json:"expectMessage"`
		ErrorCode      *string            `json:"errorCode"`
		ErrorMessage   *string            `json:"errorMessage"`
		Details        *datatypes.JSONMap `json:"details"`
		CommandText    *string            `json:"commandText"`
		ScriptText     *string            `json:"scriptText"`
		ShellPath      *string            `json:"shellPath"`
		WorkingDir     *string            `json:"workingDir"`
		ExitCode       *int64             `json:"exitCode"`
	}
	err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(runDBID)).
		Select(
			db.JobRunStep.ID, db.JobRunStep.StepOrder, db.JobRunStep.StepName, db.JobRunStep.Status,
			db.JobRunStep.StartedAt, db.JobRunStep.FinishedAt, db.JobRunStep.SqlText, db.JobRunStep.TimeoutSeconds,
			db.JobRunStep.RowsCount, db.JobRunStep.RowsAffected, db.JobRunStep.Expectation, db.JobRunStep.ExpectOk,
			db.JobRunStep.ExpectMessage, db.JobRunStep.ErrorCode, db.JobRunStep.ErrorMessage, db.JobRunStep.Details,
			db.JobRunStep.CommandText, db.JobRunStep.ScriptText, db.JobRunStep.ShellPath, db.JobRunStep.WorkingDir, db.JobRunStep.ExitCode,
		).Order(db.JobRunStep.StepOrder.Asc()).Scan(&steps)
	if err != nil {
		return nil, err
	}
	stepIDs := make([]int64, 0, len(steps))
	for _, s := range steps {
		stepIDs = append(stepIDs, s.ID)
	}
	shellIOByStep, _ := loadShellIO(stepIDs)
	webtaskIOByStep, _ := loadWebtaskIO(stepIDs)
	stepsResp := make([]gin.H, 0, len(steps))
	for _, s := range steps {
		m := gin.H{
			"id":             s.ID,
			"stepOrder":      s.StepOrder,
			"stepName":       s.StepName,
			"status":         s.Status,
			"startedAt":      s.StartedAt,
			"finishedAt":     s.FinishedAt,
			"sqlText":        s.SQLText,
			"timeoutSeconds": s.TimeoutSeconds,
			"rowsCount":      s.RowsCount,
			"rowsAffected":   s.RowsAffected,
			"expectation":    s.Expectation,
			"expectOk":       s.ExpectOK,
			"expectMessage":  s.ExpectMessage,
			"errorCode":      s.ErrorCode,
			"errorMessage":   s.ErrorMessage,
			"details":        s.Details,
			"commandText":    s.CommandText,
			"scriptText":     s.ScriptText,
			"shellPath":      s.ShellPath,
			"workingDir":     s.WorkingDir,
			"exitCode":       s.ExitCode,
		}
		if io, ok := shellIOByStep[s.ID]; ok {
			for k, v := range io {
				m[k] = v
			}
		}
		if io, ok := webtaskIOByStep[s.ID]; ok {
			for k, v := range io {
				m[k] = v
			}
		}
		stepsResp = append(stepsResp, m)
	}
	return stepsResp, nil
}

func loadShellIO(stepIDs []int64) (map[int64]gin.H, error) {
	type shellIORow struct {
		StepID           int64   `json:"stepId"`
		StdoutText       *string `json:"stdoutText"`
		StderrText       *string `json:"stderrText"`
		StdoutTruncated  bool    `json:"stdoutTruncated"`
		StderrTruncated  bool    `json:"stderrTruncated"`
		StdoutBytes      *int64  `json:"stdoutBytes"`
		StderrBytes      *int64  `json:"stderrBytes"`
		StdoutTotalBytes *int64  `json:"stdoutTotalBytes"`
		StderrTotalBytes *int64  `json:"stderrTotalBytes"`
	}
	out := map[int64]gin.H{}
	if len(stepIDs) == 0 {
		return out, nil
	}
	var rows []shellIORow
	if err := db.JobRunShellIo.Where(db.JobRunShellIo.StepID.In(stepIDs...)).
		Select(db.JobRunShellIo.StepID, db.JobRunShellIo.StdoutText, db.JobRunShellIo.StderrText, db.JobRunShellIo.StdoutTruncated, db.JobRunShellIo.StderrTruncated, db.JobRunShellIo.StdoutBytes, db.JobRunShellIo.StderrBytes, db.JobRunShellIo.StdoutTotalBytes, db.JobRunShellIo.StderrTotalBytes).
		Scan(&rows); err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.StepID] = gin.H{
			"stdoutText":       r.StdoutText,
			"stderrText":       r.StderrText,
			"stdoutTruncated":  r.StdoutTruncated,
			"stderrTruncated":  r.StderrTruncated,
			"stdoutBytes":      r.StdoutBytes,
			"stderrBytes":      r.StderrBytes,
			"stdoutTotalBytes": r.StdoutTotalBytes,
			"stderrTotalBytes": r.StderrTotalBytes,
		}
	}
	return out, nil
}

func loadWebtaskIO(stepIDs []int64) (map[int64]gin.H, error) {
	type webtaskIORow struct {
		StepID          int64              `json:"stepId"`
		RequestURL      *string            `json:"requestUrl"`
		RequestMethod   *string            `json:"requestMethod"`
		RequestHeaders  *datatypes.JSONMap `json:"requestHeaders"`
		RequestBody     *string            `json:"requestBody"`
		ResponseStatus  *int64             `json:"responseStatus"`
		ResponseHeaders *datatypes.JSONMap `json:"responseHeaders"`
		ResponseBody    *string            `json:"responseBody"`
		LatencyMs       *int64             `json:"latencyMs"`
	}
	out := map[int64]gin.H{}
	if len(stepIDs) == 0 {
		return out, nil
	}
	var rows []webtaskIORow
	if err := db.JobRunWebtaskIo.Where(db.JobRunWebtaskIo.StepID.In(stepIDs...)).Scan(&rows); err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.StepID] = gin.H{
			"requestUrl":      r.RequestURL,
			"requestMethod":   r.RequestMethod,
			"requestHeaders":  r.RequestHeaders,
			"requestBody":     r.RequestBody,
			"responseStatus":  r.ResponseStatus,
			"responseHeaders": r.ResponseHeaders,
			"responseBody":    r.ResponseBody,
			"latencyMs":       r.LatencyMs,
		}
	}
	return out, nil
}

func enrichRunObject(runObj gin.H, jr *models.JobRun) {
	if jr.ActionID != nil {
		if a, err := db.Action.Where(db.Action.ID.Eq(*jr.ActionID)).Select(db.Action.Name).First(); err == nil && a != nil {
			runObj["actionName"] = a.Name
		}
	}
	ck := ""
	if jr.ConnectionKind != nil {
		ck = strings.ToLower(*jr.ConnectionKind)
	}
	switch ck {
	case "database":
		if jr.ConnectionID != nil {
			if conn, err := db.DbConnection.Where(db.DbConnection.ID.Eq(*jr.ConnectionID)).Select(db.DbConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
	case "shell":
		if jr.ShellConnectionID != nil {
			if conn, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(*jr.ShellConnectionID)).Select(db.ShellConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
	case "webtask":
		if jr.WebtaskConnectionID != nil {
			if conn, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(*jr.WebtaskConnectionID)).Select(db.WebtaskConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
	default:
		if jr.ConnectionID != nil {
			if conn, err := db.DbConnection.Where(db.DbConnection.ID.Eq(*jr.ConnectionID)).Select(db.DbConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
		if runObj["connectionName"] == nil && jr.ShellConnectionID != nil {
			if conn, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(*jr.ShellConnectionID)).Select(db.ShellConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
		if runObj["connectionName"] == nil && jr.WebtaskConnectionID != nil {
			if conn, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(*jr.WebtaskConnectionID)).Select(db.WebtaskConnection.Name).First(); err == nil && conn != nil {
				runObj["connectionName"] = conn.Name
			}
		}
	}
}
