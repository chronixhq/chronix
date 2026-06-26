package notify

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	eventspkg "chronix/internal/events"
	"time"

	"github.com/dan-sherwin/go-utilities"
	"gorm.io/datatypes"
	"gorm.io/gorm/clause"
)

type NotificationSeverity string

const (
	SeverityInfo    NotificationSeverity = "info"
	SeveritySuccess NotificationSeverity = "success"
	SeverityWarning NotificationSeverity = "warning"
	SeverityError   NotificationSeverity = "error"
)

// NotificationCategory represents the general source/type of a notification.
// This is a soft enum (string-backed) to keep DB compatibility while giving typed guidance in code.
// Add new categories here as they are introduced.
// Common categories: job, system, connection, worker, backup, upgrade.
type NotificationCategory string

const (
	CategoryJob        NotificationCategory = "job"
	CategorySystem     NotificationCategory = "system"
	CategorySecurity   NotificationCategory = "security"
	CategoryConnection NotificationCategory = "connection"
	CategoryWorker     NotificationCategory = "worker"
)

type UserNotification struct {
	models.Notification
	Seen bool `json:"seen"`
}

// CreateNotification inserts a new notification and returns its id.
func CreateNotification(category NotificationCategory, severity NotificationSeverity, subject string, origin *string, dataJSON *map[string]any) (int64, error) {
	n := &models.Notification{
		CreatedAt: time.Now().UTC(),
		Category:  string(category),
		Severity:  string(severity),
		Subject:   subject,
		Origin:    origin,
		Data:      (*datatypes.JSONMap)(dataJSON),
	}
	if err := db.Notification.Create(n); err != nil {
		return 0, err
	}

	// Dispatch alerts (async/best-effort)
	go DispatchAlerts(n)

	if n.ID != nil {
		return *n.ID, nil
	}
	// Fallback: fetch last inserted id
	last, err := db.Notification.Order(db.Notification.ID.Desc()).Limit(1).First()
	if err != nil || last == nil || last.ID == nil {
		return 0, err
	}
	return *last.ID, nil
}

// AssignToAllUsers creates recipient rows for each user id. Ignores existing pairs.
func AssignToAllUsers(notificationID int64) error {
	var userIDs []int64
	err := db.CxUser.Select(db.CxUser.ID).Scan(&userIDs)
	if err != nil {
		return err
	}
	return AssignToUsers(notificationID, userIDs)
}

// AssignToUsers creates recipient rows for each user id. Ignores existing pairs.
func AssignToUsers(notificationID int64, userIDs []int64) error {
	if len(userIDs) == 0 {
		return nil
	}
	// Load the notification once to populate the SSE payload
	n, err := db.Notification.Where(db.Notification.ID.Eq(notificationID)).Take()
	if err != nil || n == nil || n.ID == nil {
		return err
	}
	wire := eventspkg.NotificationWire{
		ID:        *n.ID,
		CreatedAt: n.CreatedAt,
		Category:  n.Category,
		Severity:  n.Severity,
		Subject:   n.Subject,
		Origin:    n.Origin,
		Data:      n.Data,
		Seen:      false,
	}
	for _, uid := range userIDs {
		rec := &models.NotificationRecipient{
			NotificationID: notificationID,
			UserID:         uid,
			Seen:           false,
			DeliveredAt:    time.Now().UTC(),
		}
		if err := db.NotificationRecipient.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "notification_id"}, {Name: "user_id"}},
				DoNothing: true,
			}).
			Create(rec); err != nil {
			return err
		}
		// Emit SSE event with the full item for the specific user
		_ = eventspkg.SendUserIDEvent(uid, eventspkg.SSEEventNotification, eventspkg.NotificationEvent{ID: notificationID, Item: wire, UnseenDelta: 1})
	}
	return nil
}

// GetRecentForUser returns recent notifications for a user with seen flag.
func GetRecentForUser(userID int64, limit int) ([]UserNotification, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	type row struct {
		ID        int64
		CreatedAt time.Time
		Category  string
		Severity  string
		Subject   string
		Origin    *string
		Data      *datatypes.JSONMap
		Seen      bool
	}
	var rows []row
	err := db.Notification.
		Select(
			db.Notification.ID,
			db.Notification.CreatedAt,
			db.Notification.Category,
			db.Notification.Severity,
			db.Notification.Subject,
			db.Notification.Origin,
			db.Notification.Data,
			db.NotificationRecipient.Seen,
		).
		Join(db.NotificationRecipient, db.NotificationRecipient.NotificationID.EqCol(db.Notification.ID)).
		Where(
			db.NotificationRecipient.UserID.Eq(userID),
			db.NotificationRecipient.RemovedAt.IsNull(),
		).
		Order(db.Notification.CreatedAt.Desc()).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	out := make([]UserNotification, 0, len(rows))
	for _, r := range rows {
		out = append(out, UserNotification{
			Notification: models.Notification{
				ID:        utilities.Ptr(r.ID),
				CreatedAt: r.CreatedAt,
				Category:  r.Category,
				Severity:  r.Severity,
				Subject:   r.Subject,
				Origin:    r.Origin,
				Data:      r.Data,
			},
			Seen: r.Seen,
		})
	}
	return out, nil
}

func CountUnseenForUser(userID int64) (int64, error) {
	return db.NotificationRecipient.
		Where(
			db.NotificationRecipient.UserID.Eq(userID),
			db.NotificationRecipient.Seen.Is(false),
			db.NotificationRecipient.RemovedAt.IsNull(),
		).
		Count()
}

// MarkSeen sets seen=true for provided ids for a user.
func MarkSeen(userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.NotificationRecipient.
		Where(
			db.NotificationRecipient.UserID.Eq(userID),
			db.NotificationRecipient.NotificationID.In(ids...),
		).
		UpdateSimple(
			db.NotificationRecipient.Seen.Value(true),
			db.NotificationRecipient.SeenAt.Value(time.Now().UTC()),
		)
	return err
}

// ListForUser returns a paginated list for the notifications page.
func ListForUser(userID int64, page, pageSize int, category NotificationCategory, severity NotificationSeverity, seen *bool) ([]UserNotification, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 200 {
		pageSize = 50
	}
	base := db.Notification.
		Join(db.NotificationRecipient, db.NotificationRecipient.NotificationID.EqCol(db.Notification.ID)).
		Where(
			db.NotificationRecipient.UserID.Eq(userID),
			db.NotificationRecipient.RemovedAt.IsNull(),
		)
	if category != "" {
		base = base.Where(db.Notification.Category.Eq(string(category)))
	}
	if severity != "" {
		base = base.Where(db.Notification.Severity.Eq(string(severity)))
	}
	if seen != nil {
		base = base.Where(db.NotificationRecipient.Seen.Is(*seen))
	}
	// Count total
	total, err := base.Count()
	if err != nil {
		return nil, 0, err
	}
	// Page query
	offset := (page - 1) * pageSize
	type row struct {
		ID        int64
		CreatedAt time.Time
		Category  string
		Severity  string
		Subject   string
		Origin    *string
		Data      *datatypes.JSONMap
		Seen      bool
	}
	var rows []row
	err = base.
		Select(
			db.Notification.ID,
			db.Notification.CreatedAt,
			db.Notification.Category,
			db.Notification.Severity,
			db.Notification.Subject,
			db.Notification.Origin,
			db.Notification.Data,
			db.NotificationRecipient.Seen,
		).
		Order(db.Notification.CreatedAt.Desc()).
		Limit(pageSize).
		Offset(offset).
		Scan(&rows)
	if err != nil {
		return nil, 0, err
	}
	out := make([]UserNotification, 0, len(rows))
	for _, r := range rows {
		out = append(out, UserNotification{
			Notification: models.Notification{
				ID:        utilities.Ptr(r.ID),
				CreatedAt: r.CreatedAt,
				Category:  r.Category,
				Severity:  r.Severity,
				Subject:   r.Subject,
				Origin:    r.Origin,
				Data:      r.Data,
			},
			Seen: r.Seen,
		})
	}
	return out, total, nil
}

// MarkRemoved sets removed_at=now for provided notification ids for a user.
func MarkRemoved(userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := db.NotificationRecipient.
		Where(
			db.NotificationRecipient.UserID.Eq(userID),
			db.NotificationRecipient.NotificationID.In(ids...),
			// Only update if not already removed
			db.NotificationRecipient.RemovedAt.IsNull(),
		).
		UpdateSimple(
			db.NotificationRecipient.RemovedAt.Value(time.Now().UTC()),
		)
	return err
}
