package notify

import (
	"log/slog"
)

// TryCreateNotification is a helper to safely create and fanout a notification to all users.
// It logs failures instead of propagating errors to callers to avoid cascading issues
// in operational paths. Use this for best-effort alerts.
func TryCreateNotification(category NotificationCategory, severity NotificationSeverity, subject string, origin *string, dataJSON *map[string]any) {
	id, err := CreateNotification(category, severity, subject, origin, dataJSON)
	if err != nil {
		slog.Error("create notification", "component", "notify", "error", err)
		return
	}
	if err := AssignToAllUsers(id); err != nil {
		slog.Error("assign notification", "component", "notify", "id", id, "error", err)
	}
}
