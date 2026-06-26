package jobrun

import (
	"chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/execution"
	notifypkg "chronix/internal/notify"
	progresspkg "chronix/internal/progress"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dan-sherwin/go-utilities"
	"github.com/dan-sherwin/idgen"
	"golang.org/x/time/rate"
)

// Skeleton in-memory job run queue and runner (Phase 2 foundation).
//
// Responsibilities:
// - Accept immediate and scheduled-ready jobs via EnqueueJobRun.
// - Maintain a bounded in-memory queue (process-local) to feed a worker goroutine.
// - Update in-memory run history (job_runs_mem.go) as jobs transition states.
// - Stub out actual execution; real execution will be integrated later.
//
// Notes:
// - This is intentionally simple and in-memory. Persistence, retries, and
//   multi-worker fanout are out-of-scope for this skeleton.

type queuedJob struct {
	JobID       int64
	RunID       string
	TriggeredBy int64
	QueuedAt    time.Time
}

// activeRuns tracks per-run cancel functions for cancellation support.
var (
	activeRunsMu sync.Mutex
	activeRuns   = map[string]context.CancelFunc{}
)

var (
	activeJobRunsMu sync.Mutex
	activeJobRuns   = map[int64]int{}
)

var (
	queueOnce sync.Once
	jobCh     chan queuedJob
	// Allow configuration in future; keep simple now.
	queueSize = 256

	// Rate limiter for enqueue: default 1 token/ms, burst configurable
	enqueueLimiter *rate.Limiter
	limiterMu      sync.Mutex
	enqueueBurst   int32 = 1 // default burst size

	// Worker pool size (default 4)
	workerCount int32 = 4

	// Reusable ID generator instance (1ms pace, 8-char base36 with Feistel obfuscation)
	runIDGen *idgen.Generator
)

// SetupJobRunner initializes the job processing workers and input channels.
func SetupJobRunner(ctx context.Context) {
	queueOnce.Do(func() {
		jobCh = make(chan queuedJob, queueSize)
		// Initialize rate limiter: 1 token per millisecond, burst from setting
		burst := int(atomic.LoadInt32(&enqueueBurst))
		if burst < 1 {
			burst = 1
		}
		enqueueLimiter = rate.NewLimiter(rate.Every(time.Millisecond), burst)
		// Initialize run ID generator (1ms pace, 8-char width, default Feistel obfuscation)
		var err error
		runIDGen, err = idgen.New(
			idgen.WithEpoch(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			idgen.WithPace(time.Millisecond),
			idgen.WithWidth(8),
		)
		if err != nil {
			slog.Error("init run id generator", "component", "jobqueue", "error", err)
		}
		// Launch worker pool
		wc := GetWorkerCount()
		for i := 0; i < wc; i++ {
			go jobWorker(ctx, jobCh)
		}
		// Log that runner started
		slog.Info("job runner started", "component", "jobqueue", "burst", burst, "workers", wc)
	})
}

// EnqueueJobRun adds a job to the execution queue and returns the generated RunID.
func EnqueueJobRun(jobID int64, triggeredBy int64) (string, error) {
	if jobCh == nil {
		return "", errors.New("job runner not started")
	}

	// Verify job is not suspended
	job, err := db.Job.Where(db.Job.ID.Eq(jobID)).First()
	if err != nil {
		return "", fmt.Errorf("job not found: %w", err)
	}
	if job.Suspended {
		return "", errors.New("cannot enqueue job: job is suspended")
	}
	if !reserveJobRun(jobID) {
		return "", errors.New("cannot enqueue job: job is already queued or running")
	}

	// Throttle ingress: wait for 1 token per millisecond
	if enqueueLimiter != nil {
		_ = enqueueLimiter.Wait(context.Background())
	}
	// Generate 1ms-paced run ID (no sequence)
	runID := FormatRunID(generateRunID())
	// Record observable queued state before enqueue
	RecordJobRunQueued(jobID, runID, triggeredBy)
	// Progress reporting
	progresspkg.OnRunQueued(jobID, runID, triggeredBy)
	progresspkg.SetTriggeredBy(runID, triggeredBy)

	qj := queuedJob{JobID: jobID, RunID: runID, TriggeredBy: triggeredBy, QueuedAt: time.Now()}
	select {
	case jobCh <- qj:
		return runID, nil
	default:
		releaseJobRun(jobID)
		// Queue is full; mark as failed and return error
		RecordJobRunFinished(jobID, runID, "error", "job queue is full")
		progresspkg.OnRunFinished(jobID, runID, "error", "job queue is full")
		return "", errors.New("job queue is full")
	}
}

// generateRunID returns a unique, monotonic ID paced at 1ms per ID using a custom epoch.
// No sequence or randomness; acts as a natural regulator.
func generateRunID() int64 {
	if runIDGen != nil {
		return runIDGen.Generate()
	}
	// Fallback: use current UnixMilli as a monotonic-ish raw value
	return time.Now().UnixMilli()
}

// FormatRunID returns an 8-character base36 string via the shared generator.
func FormatRunID(id int64) string {
	if runIDGen != nil {
		return runIDGen.Format(id)
	}
	// Fallback (should not happen after runner init): base36 padded to 8
	s := strconv.FormatInt(id, 36)
	if len(s) < 8 {
		pad := make([]byte, 8-len(s))
		for i := range pad {
			pad[i] = '0'
		}
		s = string(pad) + s
	}
	return s
}

// UnformatRunID parses a formatted ID back to the raw tick using the generator.
func UnformatRunID(formatted string) (int64, error) {
	if runIDGen != nil {
		return runIDGen.Parse(formatted)
	}
	// Fallback parse base36 directly
	s := strings.TrimSpace(formatted)
	if s == "" {
		return 0, errors.New("empty run id")
	}
	v, err := strconv.ParseInt(s, 36, 64)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func jobWorker(ctx context.Context, ch <-chan queuedJob) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("job runner stopping", "component", "jobqueue")
			return
		case j := <-ch:
			func() {
				defer releaseJobRun(j.JobID)
				defer func() {
					if r := recover(); r != nil {
						slog.Error("recovered from panic in jobWorker", "job_id", j.JobID, "run_id", j.RunID, "error", r, "stack", string(debug.Stack()))
					}
				}()
				// Transition to running
				RecordJobRunStarted(j.JobID, j.RunID)
				progresspkg.OnRunStarted(j.JobID, j.RunID)
				// Create per-run cancelable context and register it
				runCtx, cancel := context.WithCancel(ctx)
				activeRunsMu.Lock()
				activeRuns[j.RunID] = cancel
				activeRunsMu.Unlock()
				// Execute job via executor
				status, msg, err := execution.ExecuteJob(runCtx, j.JobID, j.RunID)
				// Cleanup active run entry
				activeRunsMu.Lock()
				delete(activeRuns, j.RunID)
				activeRunsMu.Unlock()
				if err != nil {
					slog.Error("job execute", "component", "jobqueue", "job_id", j.JobID, "run_id", j.RunID, "error", err)
				}
				RecordJobRunFinished(j.JobID, j.RunID, status, msg)
				progresspkg.OnRunFinished(j.JobID, j.RunID, status, msg)

				// 1. Log to Activity Log (Normal activity)
				jobName := "Unknown"
				notifyOnSuccess := false
				notifyIncludeOutput := false
				if job, err := db.Job.Where(db.Job.ID.Eq(j.JobID)).First(); err == nil && job != nil {
					jobName = job.Name
					notifyOnSuccess = utilities.PtrVal(job.NotifyOnSuccess)
					notifyIncludeOutput = utilities.PtrVal(job.NotifyIncludeOutput)
				}
				activityDetails := fmt.Sprintf("Job: %s, RunID: %s, Status: %s, Message: %s", jobName, j.RunID, status, msg)
				_ = activity.RecordUserActivity(j.TriggeredBy, "Job finished", activityDetails, "", "")

				// 2. Trigger Alert Notification (Abnormal or requested activity)
				shouldAlert := (status == "error") || (status == "success" && notifyOnSuccess)
				if shouldAlert {
					severity := notifypkg.SeveritySuccess
					if status == "error" {
						severity = notifypkg.SeverityError
					}
					data := map[string]any{
						"job_id":      j.JobID,
						"job_name":    jobName,
						"run_id":      j.RunID,
						"status":      status,
						"message":     msg,
						"finished_at": time.Now().Format("2006-01-02 15:04:05"),
					}
					if notifyIncludeOutput {
						data["output"] = progresspkg.AggregateRunOutput(j.RunID)
					}
					notifypkg.TryCreateNotification(notifypkg.CategoryJob, severity, fmt.Sprintf("Job '%s' finished: %s", jobName, msg), nil, &data)
				}
			}()
		}
	}
}

func reserveJobRun(jobID int64) bool {
	activeJobRunsMu.Lock()
	defer activeJobRunsMu.Unlock()
	if activeJobRuns[jobID] > 0 {
		return false
	}
	activeJobRuns[jobID] = 1
	return true
}

func releaseJobRun(jobID int64) {
	activeJobRunsMu.Lock()
	defer activeJobRunsMu.Unlock()
	if activeJobRuns[jobID] <= 1 {
		delete(activeJobRuns, jobID)
		return
	}
	activeJobRuns[jobID]--
}

// SetJobQueueEnqueueBurst updates the enqueue burst (tokens allowed instantly) for the 1/ms limiter.
// If the runner is already started, it atomically swaps in a new limiter with the given burst.
// n <= 0 will be coerced to 1.
func SetJobQueueEnqueueBurst(n int) {
	if n < 1 {
		n = 1
	}
	atomic.StoreInt32(&enqueueBurst, int32(n))
	limiterMu.Lock()
	defer limiterMu.Unlock()
	if jobCh != nil {
		// Runner started; swap limiter
		enqueueLimiter = rate.NewLimiter(rate.Every(time.Millisecond), n)
	}
}

// GetJobQueueEnqueueBurst returns the currently configured enqueue burst size.
func GetJobQueueEnqueueBurst() int {
	return int(atomic.LoadInt32(&enqueueBurst))
}

// SetWorkerCount updates the number of job worker goroutines to launch on start.
// Must be called before SetupJobRunner to take effect.
func SetWorkerCount(n int) {
	if n < 1 {
		n = 1
	}
	atomic.StoreInt32(&workerCount, int32(n))
}

// GetWorkerCount returns the configured worker count.
func GetWorkerCount() int {
	return int(atomic.LoadInt32(&workerCount))
}

// CancelRun attempts to cancel a running job by runID.
// Returns true if a running job was found and cancellation was signaled.
func CancelRun(runID string) bool {
	activeRunsMu.Lock()
	cancel, ok := activeRuns[runID]
	activeRunsMu.Unlock()
	if ok && cancel != nil {
		cancel()
		return true
	}
	return false
}

// TimestampFromRunID decodes a formatted run ID (base36, 8 chars) back into
// the original timestamp. The returned time is in UTC and corresponds to
// epoch2025ms + rawMs, where rawMs is the millisecond offset since
// 2025-01-01T00:00:00Z encoded in the run ID.
func TimestampFromRunID(formatted string) (time.Time, error) {
	if runIDGen != nil {
		return runIDGen.TimestampFromID(formatted)
	}
	// Fallback: parse base36 and use Unix epoch (best-effort)
	raw, err := UnformatRunID(formatted)
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(raw).UTC(), nil
}

// TimestampFromRawRunID converts a raw run ID value into a UTC timestamp using generator config.
func TimestampFromRawRunID(raw int64) time.Time {
	if runIDGen != nil {
		return runIDGen.TimestampFromRaw(raw)
	}
	return time.UnixMilli(raw).UTC()
}
