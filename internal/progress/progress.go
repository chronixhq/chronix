package progress

import (
	"sync"
	"time"
)

type EventType string

const (
	EventRunQueued    EventType = "RunQueued"
	EventRunStarted   EventType = "RunStarted"
	EventStepStarted  EventType = "StepStarted"
	EventStepFinished EventType = "StepFinished"
	EventRunFinished  EventType = "RunFinished"
)

type Event struct {
	Ts        time.Time      `json:"ts"`
	Type      EventType      `json:"type"`
	StepIndex *int           `json:"stepIndex,omitempty"`
	StepName  *string        `json:"stepName,omitempty"`
	Status    *string        `json:"status,omitempty"`
	Message   *string        `json:"message,omitempty"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type Snapshot struct {
	RunID           string     `json:"runId"`
	JobID           int64      `json:"jobId"`
	Status          string     `json:"status"`
	CurrentStep     *int       `json:"currentStep,omitempty"`
	CurrentStepName *string    `json:"currentStepName,omitempty"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Message         *string    `json:"message,omitempty"`
}

type runProgressBuf struct {
	mu          sync.Mutex
	buf         []Event
	cap         int
	snap        Snapshot
	triggeredBy int64
}

var (
	progressMu         sync.Mutex
	progressByRun      = map[string]*runProgressBuf{}
	defaultProgressCap = 200
	eventBroadcaster   func(eventType string, data any) error
)

func SetBroadcaster(fn func(eventType string, data any) error) {
	eventBroadcaster = fn
}

func getOrCreateRunProgress(runID string, jobID int64) *runProgressBuf {
	progressMu.Lock()
	defer progressMu.Unlock()
	rb, ok := progressByRun[runID]
	if !ok {
		rb = &runProgressBuf{cap: defaultProgressCap}
		rb.snap = Snapshot{RunID: runID, JobID: jobID, Status: "queued", UpdatedAt: time.Now()}
		progressByRun[runID] = rb
	}
	return rb
}

func SetConfig(bufferCap int) {
	if bufferCap > 0 {
		defaultProgressCap = bufferCap
	}
}

func SetTriggeredBy(runID string, userID int64) {
	rb := getOrCreateRunProgress(runID, 0)
	rb.mu.Lock()
	rb.triggeredBy = userID
	rb.mu.Unlock()
}

func addEventLocked(rb *runProgressBuf, ev Event) {
	if rb.cap <= 0 {
		rb.cap = defaultProgressCap
	}
	if len(rb.buf) >= rb.cap {
		copy(rb.buf, rb.buf[1:])
		rb.buf[len(rb.buf)-1] = ev
	} else {
		rb.buf = append(rb.buf, ev)
	}
	rb.snap.UpdatedAt = ev.Ts
}

func broadcast(eventType string, data any) {
	if eventBroadcaster != nil {
		_ = eventBroadcaster(eventType, data)
	}
}

func OnRunQueued(jobID int64, runID string, triggeredBy int64) {
	rb := getOrCreateRunProgress(runID, jobID)
	rb.mu.Lock()
	ev := Event{Ts: time.Now(), Type: EventRunQueued}
	msg := "Run queued"
	ev.Message = &msg
	addEventLocked(rb, ev)
	rb.triggeredBy = triggeredBy
	rb.mu.Unlock()

	triggerSrc := ""
	if triggeredBy != 0 {
		triggerSrc = "manual"
	}
	RLOnRunQueued(jobID, runID, triggeredBy, triggerSrc)
	broadcast("job_progress", map[string]any{"runId": runID, "jobId": jobID, "type": string(EventRunQueued), "message": msg})
}

func OnRunStarted(jobID int64, runID string) {
	rb := getOrCreateRunProgress(runID, jobID)
	rb.mu.Lock()
	now := time.Now()
	ev := Event{Ts: now, Type: EventRunStarted}
	msg := "Run started"
	ev.Message = &msg
	addEventLocked(rb, ev)
	rb.snap.Status = "running"
	rb.snap.StartedAt = &now
	rb.mu.Unlock()

	RLOnRunStarted(jobID, runID)
	broadcast("job_progress", map[string]any{"runId": runID, "jobId": jobID, "type": string(EventRunStarted), "message": msg})
}

func OnStepStarted(jobID int64, runID string, stepIndex int, stepName string) {
	rb := getOrCreateRunProgress(runID, jobID)
	rb.mu.Lock()
	ev := Event{Ts: time.Now(), Type: EventStepStarted, StepIndex: &stepIndex, StepName: &stepName}
	addEventLocked(rb, ev)
	rb.snap.CurrentStep = &stepIndex
	rb.snap.CurrentStepName = &stepName
	rb.mu.Unlock()

	RLOnStepStarted(jobID, runID, stepIndex, stepName)
	broadcast("job_progress", map[string]any{
		"runId":     runID,
		"jobId":     jobID,
		"type":      string(EventStepStarted),
		"stepIndex": stepIndex,
		"stepName":  stepName,
	})
}

func OnStepFinished(jobID int64, runID string, stepIndex int, stepName string, status string, message string, fields map[string]any) {
	rb := getOrCreateRunProgress(runID, jobID)
	rb.mu.Lock()
	ev := Event{Ts: time.Now(), Type: EventStepFinished, StepIndex: &stepIndex, StepName: &stepName, Status: &status, Message: &message, Fields: fields}
	addEventLocked(rb, ev)
	rb.mu.Unlock()

	RLOnStepFinished(jobID, runID, stepIndex, stepName, status, message, fields)
	broadcast("job_progress", map[string]any{
		"runId":     runID,
		"jobId":     jobID,
		"type":      string(EventStepFinished),
		"stepIndex": stepIndex,
		"stepName":  stepName,
		"status":    status,
		"message":   message,
		"fields":    fields,
	})
}

func OnRunFinished(jobID int64, runID string, status string, message string) {
	rb := getOrCreateRunProgress(runID, jobID)
	rb.mu.Lock()
	ev := Event{Ts: time.Now(), Type: EventRunFinished, Status: &status, Message: &message}
	addEventLocked(rb, ev)
	rb.snap.Status = status
	rb.snap.Message = &message
	rb.mu.Unlock()

	RLOnRunFinished(jobID, runID, status, message)
	broadcast("job_finished", map[string]any{"runId": runID, "jobId": jobID, "status": status, "message": message})

	go func() {
		time.Sleep(30 * time.Minute)
		progressMu.Lock()
		delete(progressByRun, runID)
		progressMu.Unlock()
	}()
}

func GetRunProgress(runID string) (Snapshot, []Event, bool) {
	progressMu.Lock()
	rb, ok := progressByRun[runID]
	progressMu.Unlock()
	if !ok {
		return Snapshot{}, nil, false
	}
	rb.mu.Lock()
	defer rb.mu.Unlock()
	events := make([]Event, len(rb.buf))
	copy(events, rb.buf)
	return rb.snap, events, true
}
