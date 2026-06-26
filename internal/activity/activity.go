package activity

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"time"
)

// RecordUserActivity inserts a new user activity row.
// userID may be 0 for non-user admin sessions (admin code login).
// details, ip, and userAgent may be empty strings.
func RecordUserActivity(userID int64, action string, details string, ip string, userAgent string) error {
	if action == "" || db.UserActivity == nil {
		return nil
	}
	var detPtr *string
	if details != "" {
		detPtr = &details
	}
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	var uaPtr *string
	if userAgent != "" {
		uaPtr = &userAgent
	}
	ua := &models.UserActivity{
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
		Action:    action,
		Details:   detPtr,
		IP:        ipPtr,
		UserAgent: uaPtr,
	}
	return db.UserActivity.Create(ua)
}

// GetUserActivity returns recent activity for a user, most recent first.
func GetUserActivity(userID int64, limit int) ([]*models.UserActivity, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	items, err := db.UserActivity.
		Where(db.UserActivity.UserID.Eq(userID)).
		Order(db.UserActivity.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}
	return items, nil
}

// GetAllUserActivity returns recent activity for all users, most recent first.
func GetAllUserActivity(limit int) ([]*models.UserActivity, error) {
	if limit <= 0 || limit > 1000 {
		limit = 200
	}
	items, err := db.UserActivity.
		Order(db.UserActivity.CreatedAt.Desc()).
		Limit(limit).
		Find()
	if err != nil {
		return nil, err
	}
	return items, nil
}
