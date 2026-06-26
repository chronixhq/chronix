package serverruntime

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
)

func TestSPAMiddleware_InterceptsClientRouteAndServesIndex(t *testing.T) {
	// Fake embedded FS with index.html and an asset file
	spaFS = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html><body>APP</body></html>")},
		"app.js":     &fstest.MapFile{Data: []byte("console.log('ok')")},
	}

	r := gin.New()
	// Install middleware
	r.Use(SPAServeIndexIfHTML())
	// Simulate static handler for an asset path to validate pass-through
	r.GET("/app.js", func(c *gin.Context) { c.String(200, "asset") })
	// Simulate an API endpoint that must not be intercepted
	r.GET("/server/status", func(c *gin.Context) { c.JSON(200, gin.H{"status": CurrentServerStatus}) })

	// 1) Client-side route should serve index.html
	req := httptest.NewRequest("GET", "/jobs/list", nil)
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from SPA serve, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "APP") {
		t.Fatalf("expected SPA index content, got: %q", rec.Body.String())
	}

	// 2) Asset path should pass through (not intercepted)
	req2 := httptest.NewRequest("GET", "/app.js", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK || rec2.Body.String() != "asset" {
		t.Fatalf("expected asset handler, got code=%d body=%q", rec2.Code, rec2.Body.String())
	}

	// 3) API path in allowlist should not be intercepted
	req3 := httptest.NewRequest("GET", "/server/status", nil)
	rec3 := httptest.NewRecorder()
	r.ServeHTTP(rec3, req3)
	if rec3.Code != http.StatusOK {
		t.Fatalf("expected 200 from API handler, got %d", rec3.Code)
	}
	if !strings.Contains(rec3.Body.String(), "status") {
		t.Fatalf("expected JSON body with status, got %q", rec3.Body.String())
	}

	// 4) Prefix-based allowlist should not be intercepted
	r.GET("/feedback/attachments/4", func(c *gin.Context) { c.String(200, "attachment content") })
	req4 := httptest.NewRequest("GET", "/feedback/attachments/4", nil)
	req4.Header.Set("Accept", "text/html")
	req4.Header.Set("Sec-Fetch-Mode", "navigate")
	rec4 := httptest.NewRecorder()
	r.ServeHTTP(rec4, req4)
	if rec4.Code != http.StatusOK {
		t.Fatalf("expected 200 from API handler with prefix, got %d", rec4.Code)
	}
	if rec4.Body.String() != "attachment content" {
		t.Fatalf("expected attachment content, got %q", rec4.Body.String())
	}

	// Silence unused import complaints for fs at top if build tags differ
	var _ fs.FS
}
