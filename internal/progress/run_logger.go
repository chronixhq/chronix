package progress

import (
	"chronix/pkg/typeutil"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"chronix/internal/db"
	"chronix/internal/db/models"

	"gorm.io/datatypes"
	"gorm.io/gen/field"
)

var (
	runIdxMu    sync.Mutex
	runUIDToID  = map[string]int64{}
	stepKeyToID = map[string]int64{}
)

func cacheRunID(runUID string, id int64) {
	runIdxMu.Lock()
	defer runIdxMu.Unlock()
	runUIDToID[runUID] = id
}

func getCachedRunID(runUID string) (int64, bool) {
	runIdxMu.Lock()
	defer runIdxMu.Unlock()
	id, ok := runUIDToID[runUID]
	return id, ok
}

func cacheStepID(runUID string, stepIndex int, id int64) {
	runIdxMu.Lock()
	defer runIdxMu.Unlock()
	stepKeyToID[runUID+"|"+itoa(stepIndex)] = id
}

func getCachedStepID(runUID string, stepIndex int) (int64, bool) {
	runIdxMu.Lock()
	defer runIdxMu.Unlock()
	id, ok := stepKeyToID[runUID+"|"+itoa(stepIndex)]
	return id, ok
}

func itoa(i int) string { return strconv.Itoa(i) }

func RLOnRunQueued(jobID int64, runUID string, triggeredBy int64, triggerSource string) {
	if _, ok := getCachedRunID(runUID); ok {
		return
	}
	now := time.Now()
	var jobName *string
	var connID *int64
	var actionID *int64
	var connKind *string
	var shellConnID *int64
	if j, err := db.Job.Where(db.Job.ID.Eq(jobID)).First(); err == nil {
		jn := j.Name
		jobName = &jn
		connID = j.ConnectionID
		aid := j.ActionID
		actionID = &aid
		tk := j.TargetKind
		connKind = &tk
		shellConnID = j.ShellConnectionID
	}
	status := "queued"
	r := &models.JobRun{
		RunUID:            &runUID,
		JobID:             &jobID,
		JobName:           jobName,
		ConnectionID:      connID,
		ActionID:          actionID,
		ConnectionKind:    connKind,
		ShellConnectionID: shellConnID,
		QueuedAt:          &now,
		Status:            &status,
	}
	if triggeredBy != 0 {
		r.TriggeredBy = &triggeredBy
	}
	if triggerSource != "" {
		r.TriggerSource = &triggerSource
	}
	if err := db.JobRun.Create(r); err != nil {
		slog.Error("persist run queued", "error", err, "run_uid", runUID, "job_id", jobID)
		if existing, err2 := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).First(); err2 == nil && existing.ID != nil {
			cacheRunID(runUID, *existing.ID)
		}
		return
	}
	if r.ID != nil {
		cacheRunID(runUID, *r.ID)
	}
}

func RLOnRunStarted(jobID int64, runUID string) {
	now := time.Now()
	status := "running"
	if _, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).
		UpdateSimple(db.JobRun.Status.Value(status), db.JobRun.StartedAt.Value(now)); err != nil {
		slog.Error("persist run started", "error", err, "run_uid", runUID, "job_id", jobID)
		return
	}
	if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr.ID != nil {
		cacheRunID(runUID, *jr.ID)
	}
}

func RLOnStepStarted(_ int64, runUID string, stepIndex int, stepName string) {
	runID, ok := getCachedRunID(runUID)
	if !ok {
		if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr.ID != nil {
			runID = *jr.ID
			cacheRunID(runUID, runID)
		} else {
			slog.Error("missing run id for step start", "run_uid", runUID, "error", err)
			return
		}
	}
	now := time.Now()
	status := "running"
	order := int64(stepIndex + 1)
	if existing, err := db.JobRunStep.Where(db.JobRunStep.RunID.Eq(runID), db.JobRunStep.StepOrder.Eq(order)).
		Select(db.JobRunStep.ID).First(); err == nil && existing.ID != nil {
		_, _ = db.JobRunStep.Where(db.JobRunStep.ID.Eq(*existing.ID)).
			UpdateSimple(
				db.JobRunStep.StepName.Value(stepName),
				db.JobRunStep.StartedAt.Value(now),
				db.JobRunStep.Status.Value(status),
			)
		cacheStepID(runUID, stepIndex, *existing.ID)
		ev := &models.JobRunEvent{RunID: runID, StepID: existing.ID, CreatedAt: &now}
		kind := "info"
		msg := "step started"
		ev.Kind = &kind
		ev.Message = &msg
		_ = db.JobRunEvent.Create(ev)
		return
	}
	step := &models.JobRunStep{
		RunID:     runID,
		StepOrder: &order,
		StepName:  &stepName,
		StartedAt: &now,
		Status:    status,
	}
	if err := db.JobRunStep.Create(step); err != nil {
		slog.Error("persist step started", "error", err, "run_uid", runUID, "step", stepIndex)
		return
	}
	if step.ID != nil {
		cacheStepID(runUID, stepIndex, *step.ID)
	}
	ev := &models.JobRunEvent{RunID: runID, StepID: step.ID, CreatedAt: &now}
	kind := "info"
	msg := "step started"
	ev.Kind = &kind
	ev.Message = &msg
	_ = db.JobRunEvent.Create(ev)
}

func RLOnStepProgress(_ int64, runUID string, stepIndex int, _ string, message string, fields map[string]any) {
	runID, ok := getCachedRunID(runUID)
	if !ok {
		if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr.ID != nil {
			runID = *jr.ID
			cacheRunID(runUID, runID)
		} else {
			return
		}
	}
	var stepID *int64
	if id, ok := getCachedStepID(runUID, stepIndex); ok {
		stepID = &id
	}
	now := time.Now()
	kind := "progress"
	var data *datatypes.JSONMap
	if fields != nil {
		jm := datatypes.JSONMap(fields)
		data = &jm
	}
	ev := &models.JobRunEvent{RunID: runID, StepID: stepID, CreatedAt: &now, Kind: &kind}
	if message != "" {
		ev.Message = &message
	}
	if data != nil {
		ev.Data = data
	}
	if err := db.JobRunEvent.Create(ev); err != nil {
		slog.Error("persist step progress", "error", err, "run_uid", runUID)
	}
}

func RLOnStepFinished(_ int64, runUID string, stepIndex int, _ string, status string, message string, fields map[string]any) {
	runID, _ := getCachedRunID(runUID)
	var stepID *int64
	if id, ok := getCachedStepID(runUID, stepIndex); ok {
		stepID = &id
	}
	now := time.Now()
	var rowsCount *int64
	if rc, ok := fields["rows_count"]; ok {
		val := typeutil.AsInt64(rc)
		rowsCount = &val
	}
	var rowsAffected *int64
	if ra, ok := fields["rows_affected"]; ok {
		val := typeutil.AsInt64(ra)
		rowsAffected = &val
	}
	var details *datatypes.JSONMap
	if fields != nil {
		jm := datatypes.JSONMap(fields)
		details = &jm
	}
	var expectJSON *datatypes.JSONMap
	if v, ok := fields["expectation"]; ok && v != nil {
		if m, ok2 := v.(map[string]any); ok2 {
			jm := datatypes.JSONMap(m)
			expectJSON = &jm
		}
	}
	var expectOK *bool
	if v, ok := fields["expect_ok"]; ok {
		if b, ok2 := v.(bool); ok2 {
			expectOK = &b
		}
	}
	var expectMsg *string
	if v, ok := fields["expect_message"]; ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			expectMsg = &s
		}
	}
	var errCode *string
	if v, ok := fields["error_code"]; ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			errCode = &s
		}
	}
	var errMsg *string
	if v, ok := fields["error_message"]; ok {
		if s, ok2 := v.(string); ok2 && s != "" {
			errMsg = &s
		}
	}
	if status == "error" && errMsg == nil && message != "" {
		m := message
		errMsg = &m
	}

	order := int64(stepIndex + 1)
	assigns := []field.AssignExpr{
		db.JobRunStep.Status.Value(status),
		db.JobRunStep.FinishedAt.Value(now),
	}
	if rowsCount != nil {
		assigns = append(assigns, db.JobRunStep.RowsCount.Value(*rowsCount))
	}
	if rowsAffected != nil {
		assigns = append(assigns, db.JobRunStep.RowsAffected.Value(*rowsAffected))
	}
	if expectJSON != nil {
		assigns = append(assigns, db.JobRunStep.Expectation.Value(expectJSON))
	}
	if expectOK != nil {
		assigns = append(assigns, db.JobRunStep.ExpectOk.Value(*expectOK))
	}
	if expectMsg != nil {
		assigns = append(assigns, db.JobRunStep.ExpectMessage.Value(*expectMsg))
	}
	if errCode != nil {
		assigns = append(assigns, db.JobRunStep.ErrorCode.Value(*errCode))
	}
	if errMsg != nil {
		assigns = append(assigns, db.JobRunStep.ErrorMessage.Value(*errMsg))
	}
	if details != nil {
		assigns = append(assigns, db.JobRunStep.Details.Value(details))
	}
	if v, ok := fields["executed_code"]; ok {
		if s, ok2 := v.(string); ok2 {
			assigns = append(assigns, db.JobRunStep.SqlText.Value(s))
		} else if ps, ok2 := v.(*string); ok2 && ps != nil {
			assigns = append(assigns, db.JobRunStep.SqlText.Value(*ps))
		}
	}
	if v, ok := fields["command_text"]; ok {
		if s, ok2 := v.(string); ok2 {
			assigns = append(assigns, db.JobRunStep.CommandText.Value(s))
		} else if ps, ok2 := v.(*string); ok2 && ps != nil {
			assigns = append(assigns, db.JobRunStep.CommandText.Value(*ps))
		}
	}
	if v, ok := fields["script_text"]; ok {
		if s, ok2 := v.(string); ok2 {
			assigns = append(assigns, db.JobRunStep.ScriptText.Value(s))
		} else if ps, ok2 := v.(*string); ok2 && ps != nil {
			assigns = append(assigns, db.JobRunStep.ScriptText.Value(*ps))
		}
	}
	if v, ok := fields["shell_path"]; ok {
		if s, ok2 := v.(string); ok2 {
			assigns = append(assigns, db.JobRunStep.ShellPath.Value(s))
		} else if ps, ok2 := v.(*string); ok2 && ps != nil {
			assigns = append(assigns, db.JobRunStep.ShellPath.Value(*ps))
		}
	}
	if v, ok := fields["working_dir"]; ok {
		if s, ok2 := v.(string); ok2 {
			assigns = append(assigns, db.JobRunStep.WorkingDir.Value(s))
		} else if ps, ok2 := v.(*string); ok2 && ps != nil {
			assigns = append(assigns, db.JobRunStep.WorkingDir.Value(*ps))
		}
	}
	_, _ = db.JobRunStep.Where(db.JobRunStep.RunID.Eq(runID), db.JobRunStep.StepOrder.Eq(order)).UpdateSimple(assigns...)
	kind := "info"
	ev := &models.JobRunEvent{RunID: runID, StepID: stepID, CreatedAt: &now, Kind: &kind}
	if message != "" {
		ev.Message = &message
	}
	_ = db.JobRunEvent.Create(ev)
}

func RLOnRunFinished(_ int64, runUID string, status string, message string) {
	now := time.Now()
	_, _ = db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).
		UpdateSimple(db.JobRun.Status.Value(status), db.JobRun.FinishedAt.Value(now), db.JobRun.Message.Value(message))

	if jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runUID)).Select(db.JobRun.ID).First(); err == nil && jr.ID != nil {
		runID := *jr.ID
		var totalAffected int64
		if err := db.DB.Table(models.TableNameJobRunStep).
			Select("COALESCE(SUM(rows_affected), 0)").
			Where("run_id = ?", runID).
			Scan(&totalAffected).Error; err == nil {
			_, _ = db.JobRun.Where(db.JobRun.ID.Eq(runID)).UpdateSimple(db.JobRun.RowsAffected.Value(totalAffected))
		}
		kind := "info"
		ev := &models.JobRunEvent{RunID: runID, CreatedAt: &now, Kind: &kind, Message: &message}
		_ = db.JobRunEvent.Create(ev)
	}
}
