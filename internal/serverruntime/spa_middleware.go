package serverruntime

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SPAServeIndexIfHTML is a middleware that serves the SPA index.html for top-level
// browser navigations to non-asset paths that do not correspond to a real file
// in the embedded dist/ filesystem. This enables HTML5 pushState client routing
// (e.g., /jobs/list) while preserving API behavior for XHR/JSON requests.
func SPAServeIndexIfHTML() gin.HandlerFunc {
	// Wrap the embedded spaFS in an http.FileSystem for Open/Stat checks
	fsys := http.FS(spaFS)

	// Small allowlist of GET endpoints that should never be intercepted by SPA fallback
	alwaysAllowAPI := map[string]struct{}{
		"/sse":           {},
		"/logout":        {},
		"/server/status": {},
		"/healthz":       {},
		"/ready":         {},
	}

	// Prefix-based allowlist for API routes that should never be intercepted by SPA fallback
	// (e.g. file downloads, reports, attachments)
	alwaysAllowPrefixes := []string{
		"/feedback/attachments/",
		"/activity/export",
		"/agent/update/",
	}

	return func(c *gin.Context) {
		// Only consider GET requests
		if c.Request.Method != http.MethodGet {
			return
		}

		p := c.Request.URL.Path
		if _, ok := alwaysAllowAPI[p]; ok {
			return
		}

		for _, prefix := range alwaysAllowPrefixes {
			if strings.HasPrefix(p, prefix) {
				return
			}
		}

		// If request targets a real file inside embedded dist, let static handler serve it
		if f, err := fsys.Open(p); err == nil {
			// Check if it's a directory; if it's a file, let static serve; if dir, fall through to SPA
			if info, statErr := f.Stat(); statErr == nil && !info.IsDir() {
				_ = f.Close()
				return
			}
			_ = f.Close()
		}

		// Detect top-level browser navigations more reliably than only using Accept
		secMode := c.GetHeader("Sec-Fetch-Mode")
		secDest := c.GetHeader("Sec-Fetch-Dest")
		isNavigate := strings.EqualFold(secMode, "navigate") || strings.EqualFold(secDest, "document")

		accept := c.GetHeader("Accept")
		wantsHTML := strings.Contains(accept, "text/html") && !strings.Contains(accept, "application/json")

		// Heuristic: paths without an extension are typically client routes (e.g., /jobs/list)
		// We won't rely on this alone, but it helps reduce accidental interception.
		hasExt := path.Ext(p) != ""

		if isNavigate || (wantsHTML && !hasExt) {
			// Serve the SPA shell
			f, err := fsys.Open("/index.html")
			if err != nil {
				// As a fallback, attempt without leading slash (some FS impls differ)
				f, err = fsys.Open("index.html")
			}
			if err != nil {
				slog.Warn("spa middleware: index.html not found", "error", err)
				// Do not abort; let other handlers attempt to serve (may result in 404)
				return
			}
			defer func() {
				_ = f.Close()
			}()

			// Try to get a stable mod time for better cache semantics
			var modTime time.Time
			if info, err := f.(fs.File).Stat(); err == nil {
				modTime = info.ModTime()
			}
			if modTime.IsZero() {
				modTime = time.Now()
			}

			// Serve with a correct name so Content-Type is text/html
			http.ServeContent(c.Writer, c.Request, "index.html", modTime, f)
			c.Abort()
			return
		}
	}
}
