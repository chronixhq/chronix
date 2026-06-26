package updater

import (
	"fmt"
	"log/slog"
	"strings"
	"time"
)

func StartBackgroundUpdater(currentVersion string) {
	if !Enabled {
		return
	}

	go func() {
		ticker := time.NewTicker(CheckInterval)
		defer ticker.Stop()

		CheckAndApply(currentVersion)
		for range ticker.C {
			CheckAndApply(currentVersion)
		}
	}()
}

func CheckAndApply(currentVersion string) {
	if !Enabled {
		return
	}

	manifest, serverAvailable, err := CheckForUpdates(currentVersion)
	if err != nil {
		slog.Error("background update check", slog.String("error", err.Error()))
		return
	}

	if serverAvailable && Mode == "automatic" {
		if isInUpdateWindow(WindowStart) {
			if err := ApplyUpdate(manifest, true, currentVersion); err != nil {
				slog.Error("apply automatic update", slog.String("error", err.Error()))
			}
		} else {
			slog.Debug("Skipping automatic update: outside update window", slog.String("window", WindowStart))
		}
	}

	if AgentEnabled && AgentMode == "automatic" {
		if isInUpdateWindow(AgentWindowStart) {
			slog.Info("Triggering automatic agent updates", slog.String("version", manifest.Agent.Version))
			if TriggerAgentUpdate != nil {
				TriggerAgentUpdate(manifest.Agent.Version)
			}
		} else {
			slog.Debug("Skipping automatic agent updates: outside update window", slog.String("window", AgentWindowStart))
		}
	}
}

func isInUpdateWindow(windowStartStr string) bool {
	if windowStartStr == "" {
		return true
	}

	now := time.Now()
	parts := strings.Split(windowStartStr, ":")
	if len(parts) != 2 {
		return true
	}

	var hour, minute int
	_, err := fmt.Sscanf(parts[0], "%d", &hour)
	if err != nil {
		return true
	}
	_, err = fmt.Sscanf(parts[1], "%d", &minute)
	if err != nil {
		return true
	}

	windowStart := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
	windowEnd := windowStart.Add(time.Hour)
	return (now.After(windowStart) || now.Equal(windowStart)) && now.Before(windowEnd)
}
