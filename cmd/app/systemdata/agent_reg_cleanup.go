package systemdata

import (
	"chronix/internal/db"
	"log/slog"
	"time"
)

// StartAgentRegistrationCleanup launches a background ticker that periodically purges
// expired agent registration requests to keep the table tidy.
// Policy (MVP): delete rows with status="expired" and ExpiresAt older than a
// small grace window (5 minutes). Runs every hour.
func StartAgentRegistrationCleanup() {
	slog.Info("starting agent registration cleanup", "component", "agent-reg-cleaner", "op", "start")
	go func() {
		// Do an initial run a short time after startup to avoid startup thundering herd
		time.Sleep(15 * time.Second)
		cleanupOnce()
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			cleanupOnce()
		}
	}()
}

func cleanupOnce() {
	now := time.Now()
	grace := now.Add(-5 * time.Minute)
	info, err := db.AgentRegistrationRequest.Where(
		db.AgentRegistrationRequest.Status.Eq("expired"),
		db.AgentRegistrationRequest.ExpiresAt.Lt(grace),
	).Delete()
	if err != nil {
		slog.Error("agent reg cleanup failed", "component", "agent-reg-cleaner", "op", "cleanup", "error", err)
		return
	}
	if info.RowsAffected > 0 {
		slog.Info("agent reg cleanup", "component", "agent-reg-cleaner", "op", "cleanup", "rows", info.RowsAffected)
	}
}
