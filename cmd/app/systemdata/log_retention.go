package systemdata

import (
	"chronix/internal/db"
	"log/slog"
	"time"
)

const defaultLogRetentionDays = 365

// StartLogRetentionWorker launches a background ticker that periodically purges
// old logs based on the configured retention policy. Runs once per day.
func StartLogRetentionWorker() {
	slog.Info("starting log retention worker", "component", "retention-worker", "op", "start")
	go func() {
		// Do an initial run shortly after startup
		time.Sleep(30 * time.Second)
		purgeOldLogs()

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			purgeOldLogs()
		}
	}()
}

func purgeOldLogs() {
	cutoff := time.Now().UTC().AddDate(0, 0, -defaultLogRetentionDays)

	if info, err := db.JobRun.Where(db.JobRun.QueuedAt.Lt(cutoff)).Delete(); err != nil {
		slog.Error("failed to purge old job runs", "error", err)
	} else if info.RowsAffected > 0 {
		slog.Info("purged old job runs", "count", info.RowsAffected, "cutoff", cutoff)
	}

	if info, err := db.UserActivity.Where(db.UserActivity.CreatedAt.Lt(cutoff)).Delete(); err != nil {
		slog.Error("failed to purge old user activity", "error", err)
	} else if info.RowsAffected > 0 {
		slog.Info("purged old user activity", "count", info.RowsAffected, "cutoff", cutoff)
	}
}
