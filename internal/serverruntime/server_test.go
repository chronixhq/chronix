package serverruntime

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetServerStatus_AnonymousRouter(t *testing.T) {
	// Use test mode to avoid noisy output
	gin.SetMode(gin.TestMode)
	CurrentServerStatus = StatusActive
	eng := gin.New()
	AnonymousServerRouter(eng)

	req := httptest.NewRequest("GET", "/server/status", nil)
	rec := httptest.NewRecorder()
	eng.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "\"status\":\"active\"") {
		t.Fatalf("expected status=active in body, got %q", rec.Body.String())
	}
}
