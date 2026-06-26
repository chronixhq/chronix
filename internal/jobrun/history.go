package jobrun

import (
	"chronix/internal/db"
	notifypkg "chronix/internal/notify"
	"fmt"
	"sync"
	"time"
)

// JobRunRecord represents a single execution (or queued attempt) of a job.
// This Phase 1 implementation keeps records in-memory only (not persisted).
// It is sufficient for a basic UI history view until a durable store is added.
type JobRunRecord struct {
	RunID       string     `json:"runId"`
	JobID       int64      `json:"job_id"`
	Status      string     `json:"status"` // queued | running | success | error
	QueuedAt    time.Time  `json:"queued_at"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	TriggeredBy *int64     `json:"triggered_by"`
	Message     *string    `json:"message"`
}

var (
	jobRunsMu     sync.Mutex
	jobRunsByJob  = map[int64][]JobRunRecord{}
	maxRunsPerJob = 100 // keep a small ring per job to bound memory
	// consecutiveFailures tracks consecutive failed runs per job for alerting
	consecutiveFailures = map[int64]int{}
)

// RecordJobRunQueued appends a queued record for a job.
func RecordJobRunQueued(jobID int64, runID string, triggeredBy int64) {
	jobRunsMu.Lock()
	defer jobRunsMu.Unlock()
	rec := JobRunRecord{
		RunID:    runID,
		JobID:    jobID,
		Status:   "queued",
		QueuedAt: time.Now(),
	}
	if triggeredBy != 0 {
		rec.TriggeredBy = &triggeredBy
	}
	arr := append([]JobRunRecord{rec}, jobRunsByJob[jobID]...)
	if len(arr) > maxRunsPerJob {
		arr = arr[:maxRunsPerJob]
	}
	jobRunsByJob[jobID] = arr
}

// RecordJobRunStarted marks a previously queued run as running and sets StartedAt.
func RecordJobRunStarted(jobID int64, runID string) {
	jobRunsMu.Lock()
	defer jobRunsMu.Unlock()
	now := time.Now()
	arr := jobRunsByJob[jobID]
	for i := range arr {
		if arr[i].RunID == runID {
			arr[i].Status = "running"
			arr[i].StartedAt = &now
			break
		}
	}
	jobRunsByJob[jobID] = arr
}

// RecordJobRunFinished marks a run as finished with the provided status and message.
func RecordJobRunFinished(jobID int64, runID string, status string, message string) {
	jobRunsMu.Lock()
	defer jobRunsMu.Unlock()
	now := time.Now()
	arr := jobRunsByJob[jobID]
	for i := range arr {
		if arr[i].RunID == runID {
			arr[i].Status = status
			arr[i].FinishedAt = &now
			if message != "" {
				arr[i].Message = &message
			}
			break
		}
	}
	jobRunsByJob[jobID] = arr

	// Update consecutive failure counters and emit alerts as needed
	prev := consecutiveFailures[jobID]
	switch status {
	case "error":
		cur := prev + 1
		consecutiveFailures[jobID] = cur
		th := notifypkg.GetJobConsecutiveFailuresError()
		if th > 0 && cur == th {
			data := map[string]any{"job_id": jobID, "consecutive_failures": cur, "last_message": message}
			subject := "Job consecutive failures threshold reached"
			if job, err := db.Job.Where(db.Job.ID.Eq(jobID)).First(); err == nil && job != nil {
				data["job_name"] = job.Name
				subject = fmt.Sprintf("Job '%s' consecutive failures threshold reached", job.Name)
			}
			notifypkg.TryCreateNotification(notifypkg.CategoryJob, notifypkg.SeverityError, subject, nil, &data)
		}
	case "success":
		if prev > 0 {
			data := map[string]any{"job_id": jobID, "prev_consecutive_failures": prev}
			subject := "Job recovered after failures"
			if job, err := db.Job.Where(db.Job.ID.Eq(jobID)).First(); err == nil && job != nil {
				data["job_name"] = job.Name
				subject = fmt.Sprintf("Job '%s' recovered after failures", job.Name)
			}
			notifypkg.TryCreateNotification(notifypkg.CategoryJob, notifypkg.SeveritySuccess, subject, nil, &data)
		}
		consecutiveFailures[jobID] = 0
	}
}
