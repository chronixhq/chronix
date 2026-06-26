package cxrestapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	serverruntime "chronix/internal/serverruntime"
	"github.com/gin-gonic/gin"
)

func TestAnonymousRoutes_MountsServerStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	// Mount only the anonymous server router to avoid agent WS setup in this unit test
	serverruntime.AnonymousServerRouter(eng)

	req := httptest.NewRequest(http.MethodGet, "/server/status", nil)
	rec := httptest.NewRecorder()
	eng.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", rec.Code)
	}
}

func TestAuthenticatedRoutes_CheckAuthRequiresCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	authenticatedRoutes(eng)

	req := httptest.NewRequest(http.MethodGet, "/checkauth", nil)
	rec := httptest.NewRecorder()
	eng.ServeHTTP(rec, req)
	if rec.Code == http.StatusOK {
		t.Fatalf("expected unauthorized without cookie; got %d", rec.Code)
	}
}

func TestInitialize_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	// Mount only the handler without auth middleware to hit JSON validation path
	eng.POST("/initialize", initialize)

	req := httptest.NewRequest(http.MethodPost, "/initialize", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	eng.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("want 400 for invalid JSON, got %d", rec.Code)
	}
}
