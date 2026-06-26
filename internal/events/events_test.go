package events

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewSession_MissingIdentityNoPanic(_ *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/sse", func(c *gin.Context) {
		NewSession(c, "", 0, false) // should be a no-op
		c.Status(204)
	})
	req := httptest.NewRequest("GET", "/sse", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}

func TestSSEHelpers_Noops(_ *testing.T) {
	_ = BroadcastEvent(SSEEventServerStatus, "active")
	_ = BroadcastAdminEvent(SSEEventNotification, map[string]any{"x": 1})
	_ = SendUserIDEvent(123, SSEEventLogout, LogoutEvent{Reason: "test"})
	ShutdownAuthkeySession("sess-1")
	ShutdownUserSSESessions(456)
}
