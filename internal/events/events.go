package events

import (
	"strconv"
	"time"

	sse "github.com/dan-sherwin/go-sse"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type (
	SSEEventType string
)

const (
	SSEEventServerStatus      SSEEventType = "serverStatus"
	SSEEventLogout            SSEEventType = "logout"
	SSEEventUserUpdate        SSEEventType = "userUpdate"
	SSEEventNotification      SSEEventType = "notification"
	SSEEventAgentRegistration SSEEventType = "agent_registration"
	SSEEventAgentRegApproved  SSEEventType = "agent_registration_approved"
	SSEEventAgentRegDenied    SSEEventType = "agent_registration_denied"
	SSEEventAgentDeleted      SSEEventType = "agent_deleted"
	SSEEventConnectionHealth  SSEEventType = "connection_health"
)

// UserUpdateEvent is sent to a specific user when their profile is modified by an admin.
// The frontend uses this to refresh the in-memory user object and UI (avatar initials, etc.).
// If a change requires logout, a separate SSEEventLogout will also be sent with a reason.
type UserUpdateEvent struct {
	ID      int64   `json:"id"`
	Email   string  `json:"email"`
	Name    string  `json:"name"`
	Phone   *string `json:"phone,omitempty"`
	Admin   bool    `json:"admin"`
	Enabled bool    `json:"enabled"`
}

// LogoutEvent instructs the client to immediately log out and show a message.
type LogoutEvent struct {
	Reason string `json:"reason"`
}

// NotificationWire mirrors the fields needed by the UI for a notification item.
type NotificationWire struct {
	ID        int64              `json:"id"`
	CreatedAt time.Time          `json:"createdAt"`
	Category  string             `json:"category"`
	Severity  string             `json:"severity"`
	Subject   string             `json:"subject"`
	Origin    *string            `json:"origin,omitempty"`
	Data      *datatypes.JSONMap `json:"data,omitempty"`
	Seen      bool               `json:"seen"`
}

// NotificationEvent is the payload for a new notification SSE event.
// Includes the created item to allow clients to upsert without a refetch.
type NotificationEvent struct {
	ID          int64            `json:"id"`                    // kept for backwards compatibility
	Item        NotificationWire `json:"item"`                  // the new/updated item
	UnseenDelta int              `json:"unseenDelta,omitempty"` // optional delta to apply to unseen count
	// UnseenCount int           `json:"unseenCount,omitempty"` // optional authoritative count (not used currently)
}

// ---------- Adapter functions over go-sse ----------

// NewSession starts an SSE session for the authenticated user, using server-supplied uid and sessionID.
// sessionID maps to the user's auth key (so we can revoke just this session), and uid maps to role-qualified identity.
func NewSession(c *gin.Context, authKey string, userID int64, admin bool) {
	if authKey == "" || userID == 0 {
		return
	}
	uid := "user:" + strconv.FormatInt(userID, 10)
	if admin {
		uid = "admin:" + strconv.FormatInt(userID, 10)
	}
	sse.NewSessionWithUID(c, authKey, uid)
}

// BroadcastEvent sends an event to all connected sessions.
func BroadcastEvent(eventType SSEEventType, data any) error {
	return sse.BroadcastEvent(string(eventType), data)
}

// BroadcastAdminEvent sends an event to all connected admin sessions.
// It uses the active UID list to avoid querying the database; only connected admin sessions receive it.
func BroadcastAdminEvent(eventType SSEEventType, data any) error {
	uids := sse.ListActiveUIDs("admin:")
	if len(uids) == 0 {
		return nil
	}
	return sse.SendEventToUIDs(string(eventType), data, uids)
}

// SendUserIDEvent sends an event to a specific user id without requiring a full CxUser lookup.
// Since admin status is unknown, send to both possible prefixes to cover all sessions.
func SendUserIDEvent(userID int64, eventType SSEEventType, data any) error {
	uids := []string{"user:" + strconv.FormatInt(userID, 10), "admin:" + strconv.FormatInt(userID, 10)}
	return sse.SendEventToUIDs(string(eventType), data, uids)
}

// ShutdownAuthkeySession terminates exactly one session by its auth key.
func ShutdownAuthkeySession(authKey string) {
	sse.ShutdownBySessionID(authKey)
}

// ShutdownUserSSESessions terminates all sessions for the given user id (admin or regular).
func ShutdownUserSSESessions(userID int64) {
	sse.ShutdownByUIDs([]string{
		"user:" + strconv.FormatInt(userID, 10),
		"admin:" + strconv.FormatInt(userID, 10),
	})
}

// ShutdownAllSSESessions terminates all active SSE sessions for both regular users and admins.
// Returns the number of sessions targeted for shutdown (based on UIDs collected).
func ShutdownAllSSESessions() int {
	// Collect all active admin and user UIDs. Our system uses only these prefixes.
	adminUIDs := sse.ListActiveUIDs("admin:")
	userUIDs := sse.ListActiveUIDs("user:")
	if len(adminUIDs) == 0 && len(userUIDs) == 0 {
		return 0
	}
	uids := make([]string, 0, len(adminUIDs)+len(userUIDs))
	uids = append(uids, adminUIDs...)
	uids = append(uids, userUIDs...)
	sse.ShutdownByUIDs(uids)
	return len(uids)
}
