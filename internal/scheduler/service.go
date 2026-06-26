package scheduler

import (
	"chronix/internal/db"
	jobrunpkg "chronix/internal/jobrun"
	notifypkg "chronix/internal/notify"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"runtime/debug"
	"time"

	"gorm.io/datatypes"
)

// Start launches the in-process scheduler loop that enqueues due jobs at the right time.
// It periodically loads enabled jobs and their schedules from the database, computes the
// next run time using NextRunTime, and enqueues a run via jobrun.EnqueueJobRun when due.
//
// Notes:
//   - This is a minimal, single-process scheduler suitable for Phase 2.
//   - Same-job overlap is prevented by the in-process job queue; due triggers while a job is
//     already queued or running are rejected and surfaced as dispatch failures/alerts.
//   - It respects context cancellation via the provided ctx.
func Start(ctx context.Context) {
	go runLoop(ctx)
}

type jobEntry struct {
	id     int64
	name   string
	sched  []byte
	nextAt time.Time
}

func runLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("scheduler recovered from panic", "error", r, "stack", string(debug.Stack()))
			time.Sleep(time.Second)
			go runLoop(ctx)
		}
	}()
	const (
		refreshInterval = 30 * time.Second
	)
	logger := slog.With("component", "scheduler")
	entries := map[int64]*jobEntry{}
	// Initial load
	entries = loadJobs(logger, entries)
	refTicker := time.NewTicker(refreshInterval)
	defer refTicker.Stop()

	var soonestTimer *time.Timer
	defer func() {
		if soonestTimer != nil {
			soonestTimer.Stop()
		}
	}()

	for {
		// Determine the next due time across all jobs
		next, haveNext := soonest(entries)
		var dueTimerChan <-chan time.Time
		if haveNext {
			d := time.Until(next)
			if d < 0 {
				d = 0
			}
			if soonestTimer == nil {
				soonestTimer = time.NewTimer(d)
			} else {
				if !soonestTimer.Stop() {
					select {
					case <-soonestTimer.C:
					default:
					}
				}
				soonestTimer.Reset(d)
			}
			dueTimerChan = soonestTimer.C
		} else {
			if soonestTimer != nil {
				soonestTimer.Stop()
				soonestTimer = nil
			}
		}
		select {
		case <-ctx.Done():
			logger.Info("scheduler stopped")
			return
		case <-refTicker.C:
			entries = loadJobs(logger, entries)
		case <-dueTimerChan:
			now := time.Now().UTC().Truncate(time.Minute)
			// Enqueue all jobs due now or earlier (tolerate slight delays)
			for _, e := range entries {
				if !e.nextAt.IsZero() && (e.nextAt.Equal(now) || e.nextAt.Before(now)) {
					// Missed schedule window detection
					grace := notifypkg.GetJobMissGrace()
					if grace > 0 && now.Sub(e.nextAt) > grace {
						data := map[string]any{"job_id": e.id, "job_name": e.name, "schedule_time": e.nextAt, "now": now, "grace": notifypkg.GetJobMissGrace().String()}
						notifypkg.TryCreateNotification(notifypkg.CategoryJob, notifypkg.SeverityWarning, fmt.Sprintf("Job '%s' missed schedule window", e.name), nil, &data)
					}
					if _, err := jobrunpkg.EnqueueJobRun(e.id, 0); err != nil {
						logger.Error("enqueue job", "job_id", e.id, "error", err)
						// Dispatch failure alert
						data := map[string]any{"job_id": e.id, "job_name": e.name, "error": err.Error()}
						notifypkg.TryCreateNotification(notifypkg.CategoryJob, notifypkg.SeverityError, fmt.Sprintf("Job '%s' dispatch failed", e.name), nil, &data)
					} else {
						logger.Info("job enqueued", "job_id", e.id, "next_at", e.nextAt)
					}
					// Compute the following occurrence
					if t, err := NextRunTime(e.sched, now.Add(time.Minute)); err == nil {
						e.nextAt = t
					} else {
						// No next run or parse error => drop from scheduling until next refresh fills it
						e.nextAt = time.Time{}
					}
				}
			}
		}
	}
}

// soonest finds the earliest next run time among all job entries.
// Returns the time and true if at least one entry has a valid next run time.
func soonest(entries map[int64]*jobEntry) (time.Time, bool) {
	var next time.Time
	have := false
	for _, e := range entries {
		if e.nextAt.IsZero() {
			continue
		}
		if !have || e.nextAt.Before(next) {
			next = e.nextAt
			have = true
		}
	}
	return next, have
}

// loadJobs queries the database for enabled jobs and updates the entries map.
// It reuses existing entries if the schedule JSON hasn't changed to maintain scheduling continuity.
func loadJobs(logger *slog.Logger, prev map[int64]*jobEntry) map[int64]*jobEntry {
	// Prepare fresh map; reuse previous entries when schedule unchanged to keep nextAt
	res := map[int64]*jobEntry{}
	// Query enabled jobs (nil Enabled is treated as enabled=false for now)
	type row struct {
		ID           int64
		Name         string
		ScheduleJSON datatypes.JSONMap
	}
	var rows []row
	if err := db.Job.Where(db.Job.Enabled.Is(true), db.Job.Suspended.Is(false)).Select(db.Job.ID, db.Job.Name, db.Job.ScheduleJSON).Scan(&rows); err != nil {
		logger.Error("load jobs", "error", err)
		return prev
	}
	now := time.Now().UTC().Truncate(time.Minute)
	for _, r := range rows {
		b, _ := json.Marshal(r.ScheduleJSON)
		// Check if unchanged
		if old, ok := prev[r.ID]; ok {
			if string(old.sched) == string(b) && !old.nextAt.IsZero() && old.name == r.Name {
				res[r.ID] = old
				continue
			}
		}
		// Compute next run
		t, err := NextRunTime(b, now)
		if err != nil {
			// No next run or parse error: skip scheduling this job
			continue
		}
		res[r.ID] = &jobEntry{id: r.ID, name: r.Name, sched: b, nextAt: t}
	}
	return res
}
