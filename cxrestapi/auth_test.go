package cxrestapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthAdminCode_Invalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/auth/settings/:code", authAdminCode)

	req := httptest.NewRequest(http.MethodGet, "/auth/settings/WRONG", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized { // restresponse maps to 401
		t.Fatalf("want 401 for invalid code, got %d", rec.Code)
	}
}
