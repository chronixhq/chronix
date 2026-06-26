package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type resultsPageData struct {
	Limit       int
	ResultsPort int
	APIPort     int
	WebhookPort int
	Snapshot    ResultsSnapshot
}

var resultsPageTemplate = template.Must(template.New("results").Funcs(template.FuncMap{
	"truncate": func(value string, max int) string {
		if len(value) <= max {
			return value
		}
		if max <= 3 {
			return value[:max]
		}
		return value[:max-3] + "..."
	},
}).Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="15">
  <title>Chronix Tester Results</title>
  <style>
    :root {
      color-scheme: light;
      --bg: #f5f7fb;
      --card: #ffffff;
      --line: #d6dcea;
      --text: #142033;
      --muted: #54637a;
      --accent: #0d6a9f;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      padding: 24px;
      background: linear-gradient(180deg, #eef4ff 0%, var(--bg) 100%);
      color: var(--text);
      font-family: "SF Pro Text", "Segoe UI", sans-serif;
    }
    h1, h2 { margin: 0 0 12px; }
    p { color: var(--muted); }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      margin: 16px 0 24px;
    }
    .card, .panel {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 16px;
      box-shadow: 0 12px 24px rgba(20, 32, 51, 0.06);
    }
    .card {
      padding: 16px;
    }
    .count {
      display: block;
      font-size: 30px;
      font-weight: 700;
      margin-top: 8px;
    }
    .panel {
      padding: 18px;
      margin-bottom: 16px;
    }
    .row {
      padding: 12px 0;
      border-top: 1px solid var(--line);
    }
    .row:first-child { border-top: none; padding-top: 0; }
    code, pre {
      background: #eff4fb;
      border-radius: 10px;
      padding: 2px 6px;
      font-family: "SF Mono", "JetBrains Mono", monospace;
    }
    pre {
      display: block;
      padding: 12px;
      overflow: auto;
      white-space: pre-wrap;
      margin: 8px 0 0;
    }
    .meta {
      color: var(--muted);
      font-size: 14px;
    }
    .links a {
      color: var(--accent);
      text-decoration: none;
      margin-right: 12px;
    }
  </style>
</head>
<body>
  <h1>Chronix Tester Results</h1>
  <p>Auto-refreshing local fixture dashboard for Chronix execution testing.</p>
  <p class="links">
    <a href="/api/summary">Summary JSON</a>
    <a href="/api/snapshot?limit={{.Limit}}">Full Snapshot JSON</a>
    <a href="http://127.0.0.1:{{.APIPort}}/json">API Fixture</a>
    <a href="http://127.0.0.1:{{.WebhookPort}}/healthz">Webhook Health</a>
  </p>

  <div class="grid">
    <div class="card"><strong>Target Activity</strong><span class="count">{{.Snapshot.Summary.TargetActivityCount}}</span></div>
    <div class="card"><strong>API Requests</strong><span class="count">{{.Snapshot.Summary.APILogCount}}</span></div>
    <div class="card"><strong>Webhook Posts</strong><span class="count">{{.Snapshot.Summary.WebhookLogCount}}</span></div>
    <div class="card"><strong>IMAP Messages</strong><span class="count">{{.Snapshot.Summary.IMAPLogCount}}</span></div>
    <div class="card"><strong>Shell Token Runs</strong><span class="count">{{.Snapshot.Summary.ShellLogCount}}</span></div>
  </div>

  <div class="panel">
    <h2>Paths</h2>
    <div class="row"><strong>Data Dir:</strong> <code>{{.Snapshot.Summary.DataDir}}</code></div>
    <div class="row"><strong>Results DB:</strong> <code>{{.Snapshot.Summary.ResultsDB}}</code></div>
    <div class="row"><strong>Target DB:</strong> <code>{{.Snapshot.Summary.TargetDB}}</code></div>
  </div>

  <div class="panel">
    <h2>Target Activity</h2>
    {{range .Snapshot.TargetActivity}}
      <div class="row">
        <div><strong>{{.Operation}}</strong> on <code>{{.TableName}}</code></div>
        <div class="meta">{{.CreatedAt}}</div>
        {{if .OldData}}<pre>old: {{.OldData}}</pre>{{end}}
        {{if .NewData}}<pre>new: {{.NewData}}</pre>{{end}}
      </div>
    {{else}}
      <div class="row meta">No target activity logged yet.</div>
    {{end}}
  </div>

  <div class="panel">
    <h2>API Requests</h2>
    {{range .Snapshot.APILogs}}
      <div class="row">
        <div><strong>{{.Method}}</strong> <code>{{.Path}}</code></div>
        <div class="meta">{{.CreatedAt}}</div>
        {{if .Body}}<pre>{{truncate .Body 1000}}</pre>{{end}}
      </div>
    {{else}}
      <div class="row meta">No API traffic captured yet.</div>
    {{end}}
  </div>

  <div class="panel">
    <h2>Webhook Posts</h2>
    {{range .Snapshot.WebhookLogs}}
      <div class="row">
        <div><strong>{{.Method}}</strong> <code>{{.Path}}</code></div>
        <div class="meta">{{.CreatedAt}}</div>
        {{if .Body}}<pre>{{truncate .Body 1000}}</pre>{{end}}
      </div>
    {{else}}
      <div class="row meta">No webhook traffic captured yet.</div>
    {{end}}
  </div>

  <div class="panel">
    <h2>IMAP Messages</h2>
    {{range .Snapshot.IMAPLogs}}
      <div class="row">
        <div><strong>{{.Subject}}</strong></div>
        <div class="meta">from {{.FromAddr}} at {{.ReceivedAt}}</div>
      </div>
    {{else}}
      <div class="row meta">No IMAP messages captured yet.</div>
    {{end}}
  </div>

  <div class="panel">
    <h2>Shell Token Runs</h2>
    {{range .Snapshot.ShellLogs}}
      <div class="row">
        <div><strong>Args:</strong> <code>{{.Args}}</code></div>
        <div class="meta">{{.CreatedAt}}</div>
        {{if .Output}}<pre>{{.Output}}</pre>{{end}}
      </div>
    {{else}}
      <div class="row meta">No shell token runs logged yet.</div>
    {{end}}
  </div>
</body>
</html>`))

func runServices() error {
	paths, err := resolvePaths(CLI.DataDir)
	if err != nil {
		return err
	}

	store, err := OpenStore(paths)
	if err != nil {
		return err
	}
	defer store.Close()

	ctx, stop := signalContext()
	defer stop()

	resultsServer := newHTTPServer(CLI.Run.ResultsPort, newResultsRouter(store))
	apiServer := newHTTPServer(CLI.Run.APIPort, newAPIRouter(store.Results))
	webhookServer := newHTTPServer(CLI.Run.WebhookPort, newWebhookRouter(store.Results))

	errCh := make(chan error, 3)
	go serveHTTP("results-ui", resultsServer, errCh)
	go serveHTTP("api-fixture", apiServer, errCh)
	go serveHTTP("webhook-capture", webhookServer, errCh)
	go runIMAPWorker(ctx, store.Results, CLI.Run.IMAPPollInterval)

	slog.Info("chronix tester running",
		"data_dir", paths.DataDir,
		"results_url", fmt.Sprintf("http://127.0.0.1:%d", CLI.Run.ResultsPort),
		"api_url", fmt.Sprintf("http://127.0.0.1:%d", CLI.Run.APIPort),
		"webhook_url", fmt.Sprintf("http://127.0.0.1:%d", CLI.Run.WebhookPort),
	)

	select {
	case <-ctx.Done():
		slog.Info("shutdown requested")
	case err := <-errCh:
		stop()
		if err != nil {
			slog.Error("service failed", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var errs []error
	for name, srv := range map[string]*http.Server{
		"results-ui":      resultsServer,
		"api-fixture":     apiServer,
		"webhook-capture": webhookServer,
	} {
		if err := srv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs = append(errs, fmt.Errorf("%s shutdown: %w", name, err))
		}
	}

	return errors.Join(errs...)
}

func newHTTPServer(port int, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
}

func serveHTTP(name string, srv *http.Server, errCh chan<- error) {
	slog.Info("listening", "service", name, "addr", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		errCh <- fmt.Errorf("%s: %w", name, err)
	}
}

func newResultsRouter(store *Store) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		limit := queryLimit(c, 25)
		snapshot, err := store.LoadSnapshot(limit)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		if err := resultsPageTemplate.Execute(c.Writer, resultsPageData{
			Limit:       limit,
			ResultsPort: CLI.Run.ResultsPort,
			APIPort:     CLI.Run.APIPort,
			WebhookPort: CLI.Run.WebhookPort,
			Snapshot:    snapshot,
		}); err != nil {
			slog.Error("render results page", "error", err)
		}
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/summary", func(c *gin.Context) {
		snapshot, err := store.LoadSnapshot(5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, snapshot.Summary)
	})

	r.GET("/api/snapshot", func(c *gin.Context) {
		snapshot, err := store.LoadSnapshot(queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, snapshot)
	})

	r.GET("/api/target-activity", func(c *gin.Context) {
		logs, err := queryTargetActivity(store.Results, queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})

	r.GET("/api/api-logs", func(c *gin.Context) {
		logs, err := queryAPILogs(store.Results, queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})

	r.GET("/api/webhook-logs", func(c *gin.Context) {
		logs, err := queryWebhookLogs(store.Results, queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})

	r.GET("/api/imap-logs", func(c *gin.Context) {
		logs, err := queryIMAPLogs(store.Results, queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})

	r.GET("/api/shell-logs", func(c *gin.Context) {
		logs, err := queryShellLogs(store.Results, queryLimit(c, 50))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, logs)
	})

	return r
}

func newAPIRouter(db *sql.DB) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(apiLogMiddleware(db))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.Any("/token", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"token":    fmt.Sprintf("api-token-%d", time.Now().UnixNano()),
			"issuedAt": time.Now().UTC().Format(time.RFC3339Nano),
		})
	})

	r.Any("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"slideshow": gin.H{
				"title":  "Chronix Fixture API",
				"author": "chronix-tester",
				"slides": []gin.H{
					{"title": "Overview", "type": "all"},
					{"title": "Execution", "type": "observability"},
				},
			},
		})
	})

	r.Any("/html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`<!DOCTYPE html><html><body><h1>Chronix Test Fixture</h1><p>Fixture HTML for regex capture testing.</p></body></html>`))
	})

	r.Any("/get", func(c *gin.Context) {
		c.JSON(http.StatusOK, requestDescriptor(c, bodyFromContext(c)))
	})

	r.Any("/headers", func(c *gin.Context) {
		c.Header("X-Test-ID", "12345")
		c.JSON(http.StatusOK, gin.H{"message": "Check headers"})
	})

	r.Any("/response-headers", func(c *gin.Context) {
		applied := map[string]string{}
		for key, values := range c.Request.URL.Query() {
			if len(values) == 0 {
				continue
			}
			c.Header(key, values[0])
			applied[key] = values[0]
		}
		c.JSON(http.StatusOK, gin.H{"headers": applied})
	})

	r.Any("/status/:code", func(c *gin.Context) {
		code, err := strconv.Atoi(c.Param("code"))
		if err != nil || code < 100 || code > 599 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status code"})
			return
		}
		c.JSON(code, gin.H{"status": code})
	})

	r.Any("/delay/:ms", func(c *gin.Context) {
		ms, err := strconv.Atoi(c.Param("ms"))
		if err != nil || ms < 0 || ms > 10000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "delay must be between 0 and 10000 milliseconds"})
			return
		}
		timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
		defer timer.Stop()
		select {
		case <-c.Request.Context().Done():
			c.JSON(http.StatusRequestTimeout, gin.H{"error": "request canceled"})
		case <-timer.C:
			c.JSON(http.StatusOK, gin.H{"delayedMs": ms})
		}
	})

	r.Any("/echo", func(c *gin.Context) {
		body := bodyFromContext(c)
		if len(body) == 0 {
			c.JSON(http.StatusOK, requestDescriptor(c, body))
			return
		}

		contentType := c.GetHeader("Content-Type")
		if contentType == "" {
			if json.Valid(body) {
				contentType = "application/json"
			} else {
				contentType = "text/plain; charset=utf-8"
			}
		}
		c.Data(http.StatusOK, contentType, body)
	})

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "API Test response",
			"path":    c.Request.URL.RequestURI(),
			"method":  c.Request.Method,
			"query":   c.Request.URL.Query(),
		})
	})

	return r
}

func newWebhookRouter(db *sql.DB) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	handler := func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		if _, err := db.Exec(
			`INSERT INTO webhook_logs (method, path, headers, body) VALUES (?, ?, ?, ?)`,
			c.Request.Method,
			c.Request.URL.RequestURI(),
			headerJSON(c.Request.Header),
			string(body),
		); err != nil {
			slog.Warn("record webhook request", "error", err)
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"path":   c.Request.URL.RequestURI(),
		})
	}

	r.Any("/", handler)
	r.NoRoute(handler)

	return r
}

func apiLogMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		body, _ := io.ReadAll(io.LimitReader(c.Request.Body, 1<<20))
		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Set("request_body", body)
		c.Next()

		if _, err := db.Exec(
			`INSERT INTO api_logs (method, path, headers, body) VALUES (?, ?, ?, ?)`,
			c.Request.Method,
			c.Request.URL.RequestURI(),
			headerJSON(c.Request.Header),
			string(body),
		); err != nil {
			slog.Warn("record api request", "error", err)
		}
	}
}

func bodyFromContext(c *gin.Context) []byte {
	if raw, ok := c.Get("request_body"); ok {
		if body, ok := raw.([]byte); ok {
			return body
		}
	}
	return nil
}

func requestDescriptor(c *gin.Context, body []byte) gin.H {
	return gin.H{
		"method":  c.Request.Method,
		"path":    c.Request.URL.Path,
		"query":   c.Request.URL.Query(),
		"headers": c.Request.Header,
		"body":    string(body),
	}
}

func headerJSON(header http.Header) string {
	data, err := json.Marshal(header)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func queryLimit(c *gin.Context, fallback int) int {
	raw := c.Query("limit")
	if raw == "" {
		return clampLimit(fallback, fallback)
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return clampLimit(fallback, fallback)
	}
	return clampLimit(value, fallback)
}
