package cxrestapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cxsettingspkg "chronix/internal/cxsettings"

	"github.com/gin-gonic/gin"
)

func TestPostRestartNetworkServers_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ptrBool := func(v bool) *bool { return &v }
	// Ensure the background goroutine doesn't try to start any listeners during this test.
	cxsettingspkg.CxSettings.HTTPEnabled = ptrBool(false)
	cxsettingspkg.CxSettings.HTTPSEnabled = ptrBool(false)
	cxsettingspkg.CxSettings.AgentEnabled = ptrBool(false)

	r := gin.New()
	r.POST("/settings/restart-network", postRestartNetworkServers)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/settings/restart-network", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d; body=%s", http.StatusOK, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "restarting-network") {
		t.Fatalf("expected body to contain %q, got %s", "restarting-network", w.Body.String())
	}
}
