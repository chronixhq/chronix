package cxrestapi

import (
	cxsettingspkg "chronix/internal/cxsettings"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAgentsRegisterBegin_AgentDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	old := cxsettingspkg.CxSettings.AgentEnabled
	b := false
	cxsettingspkg.CxSettings.AgentEnabled = &b
	t.Cleanup(func() {
		cxsettingspkg.CxSettings.AgentEnabled = old
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/agent/register", strings.NewReader(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	agentsRegisterBegin(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d (body=%s)", http.StatusForbidden, w.Code, w.Body.String())
	}
	if !strings.Contains(strings.ToLower(w.Body.String()), "agent connections are disabled") {
		t.Fatalf("expected error message about agent connections disabled, got body=%s", w.Body.String())
	}
}
