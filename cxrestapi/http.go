package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	cxsettingspkg "chronix/internal/cxsettings"
	cxuserpkg "chronix/internal/cxuser"
	eventspkg "chronix/internal/events"
	serverruntime "chronix/internal/serverruntime"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server"
	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

type (
	HTTPCommandDef struct {
		ForceHTTPPort  int32 `name:"force-http-port" help:"Override settings and force HTTP port" default:"0"`
		ForceHTTPSPort int32 `name:"force-https-port" help:"Override settings and force HTTPS port" default:"0"`
		ForceAgentPort int32 `name:"force-agent-port" help:"Override settings and force Agent port" default:"0"`
		DisableHTTP    bool  `name:"disable-http" help:"Disable HTTP listener"`
		DisableHTTPS   bool  `name:"disable-https" help:"Disable HTTPS listener"`
	}
)

const (
	DefaultHTTPPort  = 5170
	DefaultHTTPSPort = 5171
	DefaultAgentPort = 5172
)

var (
	CLIHTTPConfig         HTTPCommandDef
	agentListeningAddress = fmt.Sprintf("0.0.0.0:%d", DefaultAgentPort)
)

func Start() {
	defer recoverHTTPStartPanic()
	cfg := currentListenerConfig()
	registerHTTPRouters()
	startConfiguredListeners(cfg, "Starting")
}

// ShutdownServer performs a graceful shutdown of server components (SSE and HTTP/HTTPS listeners).
func ShutdownServer() {
	start := time.Now()
	n := eventspkg.ShutdownAllSSESessions()
	slog.Info("terminated SSE sessions", slog.Int("count", n), slog.Duration("took", time.Since(start)))

	if err := restapi.ShutdownHttpServerWithTimeout(time.Second*5, true); err != nil {
		slog.Warn("shutdown http", "error", err)
	}
	if err := restapi.ShutdownHttpsServerWithTimeout(time.Second*5, true); err != nil {
		slog.Warn("shutdown https", "error", err)
	}
	// small pause to allow sockets to close fully
	time.Sleep(300 * time.Millisecond)
}

func anonymousRoutes(app *gin.Engine) {
	// SPA fallback for top-level navigations before any API handlers
	app.Use(serverruntime.SPAServeIndexIfHTML())

	authRouter(app)
	agentWSRouter(app)
	agentsAnonymousRouter(app)
	agentSelfRouter(app)
	agentUpdaterRouter(app)
	serverruntime.AnonymousServerRouter(app)
}

func authenticatedRoutes(app *gin.Engine) {
	app.Use(AuthGinMiddleware())
	// After auth, still apply SPA fallback so deep-linking under authenticated routes works
	app.Use(serverruntime.SPAServeIndexIfHTML())

	app.GET("/checkauth", func(c *gin.Context) { restresponse.RestSuccessNoContent(c) })
	app.POST("/initialize", adminFunc(initialize))
	app.POST("/initialize/finalize", adminFunc(finalizeInitialization))
	app.GET("/logout", logout)
	app.GET("/sse", func(c *gin.Context) {
		user := userFromGinContext(c)
		eventspkg.NewSession(c, user.AuthKey, user.ID, user.Admin)
	})
	// Dashboard endpoints
	dashboardRouter(app)
	// Unified activity endpoint
	activityRouter(app)
	agentsAdminRouter(app)
	settingsRouter(app)
	webhooksRouter(app)
	updaterRouter(app)
	userRouter(app)
	notificationsRouter(app)
	connectionsRouter(app)
	// Additive routers for Shell and WebTask features (minimal churn; DB routes unchanged)
	shellConnectionsRouter(app)
	webtaskConnectionsRouter(app)
	shellActionsRouter(app)
	webtaskActionsRouter(app)
	actionsRouter(app)
	jobsRouter(app)
	runsRouter(app)
	helpRouter(app)
	feedbackRouter(app)
}

// GetServerUIURL constructs the best-guess URL for the management UI.
func GetServerUIURL() string {
	if cxsettingspkg.CxSettings.ServerURL != nil && *cxsettingspkg.CxSettings.ServerURL != "" {
		return *cxsettingspkg.CxSettings.ServerURL
	}

	ip := getPreferredIP()

	httpsPort := DefaultHTTPSPort
	if cxsettingspkg.CxSettings.HTTPSPort != nil && *cxsettingspkg.CxSettings.HTTPSPort > 0 {
		httpsPort = int(*cxsettingspkg.CxSettings.HTTPSPort)
	}
	if CLIHTTPConfig.ForceHTTPSPort > 0 {
		httpsPort = int(CLIHTTPConfig.ForceHTTPSPort)
	}

	httpPort := DefaultHTTPPort
	if cxsettingspkg.CxSettings.HTTPPort != nil && *cxsettingspkg.CxSettings.HTTPPort > 0 {
		httpPort = int(*cxsettingspkg.CxSettings.HTTPPort)
	}
	if CLIHTTPConfig.ForceHTTPPort > 0 {
		httpPort = int(CLIHTTPConfig.ForceHTTPPort)
	}

	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		httpsEnabled := !CLIHTTPConfig.DisableHTTPS
		httpEnabled := !CLIHTTPConfig.DisableHTTP

		if httpsEnabled && httpEnabled {
			return fmt.Sprintf("https://%s:%d or http://%s:%d", ip, httpsPort, ip, httpPort)
		} else if httpsEnabled {
			return fmt.Sprintf("https://%s:%d", ip, httpsPort)
		} else if httpEnabled {
			return fmt.Sprintf("http://%s:%d", ip, httpPort)
		}
		return "No UI listeners enabled"
	}

	// HTTPS (UI only)
	httpsEnabled := true
	if cxsettingspkg.CxSettings.HTTPSEnabled != nil {
		httpsEnabled = *cxsettingspkg.CxSettings.HTTPSEnabled
	}
	if CLIHTTPConfig.DisableHTTPS {
		httpsEnabled = false
	}

	httpEnabled := true
	if cxsettingspkg.CxSettings.HTTPEnabled != nil {
		httpEnabled = *cxsettingspkg.CxSettings.HTTPEnabled
	}
	if CLIHTTPConfig.DisableHTTP {
		httpEnabled = false
	}

	if httpsEnabled {
		return fmt.Sprintf("https://%s:%d", ip, httpsPort)
	}
	if httpEnabled {
		return fmt.Sprintf("http://%s:%d", ip, httpPort)
	}

	return "No UI listeners enabled"
}

func getPreferredIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer func() { _ = conn.Close() }()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func initialize(c *gin.Context) {
	var initData struct {
		cxuserpkg.CxUser
		ServerURL string `json:"serverUrl"`
	}
	if err := c.BindJSON(&initData); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	initData.Admin = true
	err := initData.Save()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving user", err.Error())
		return
	}
	cxsettingspkg.CxSettings.ServerURL = &initData.ServerURL
	// Ensure HTTP and HTTPS are explicitly enabled and ports are set upon initialization
	if cxsettingspkg.CxSettings.HTTPEnabled == nil {
		cxsettingspkg.CxSettings.HTTPEnabled = utilities.Ptr(true)
	}
	if cxsettingspkg.CxSettings.HTTPSEnabled == nil {
		cxsettingspkg.CxSettings.HTTPSEnabled = utilities.Ptr(true)
	}
	if cxsettingspkg.CxSettings.AgentEnabled == nil {
		cxsettingspkg.CxSettings.AgentEnabled = utilities.Ptr(true)
	}
	if cxsettingspkg.CxSettings.HTTPPort == nil || *cxsettingspkg.CxSettings.HTTPPort <= 0 {
		cxsettingspkg.CxSettings.HTTPPort = utilities.Ptr(int64(DefaultHTTPPort))
	}
	if cxsettingspkg.CxSettings.HTTPSPort == nil || *cxsettingspkg.CxSettings.HTTPSPort <= 0 {
		cxsettingspkg.CxSettings.HTTPSPort = utilities.Ptr(int64(DefaultHTTPSPort))
	}
	if cxsettingspkg.CxSettings.AgentPort == nil || *cxsettingspkg.CxSettings.AgentPort <= 0 {
		cxsettingspkg.CxSettings.AgentPort = utilities.Ptr(int64(DefaultAgentPort))
	}

	// Default updater settings: enabled and notify only
	if cxsettingspkg.CxSettings.UpdaterEnabled == nil {
		cxsettingspkg.CxSettings.UpdaterEnabled = utilities.Ptr(true)
	}
	if cxsettingspkg.CxSettings.UpdaterMode == nil {
		cxsettingspkg.CxSettings.UpdaterMode = utilities.Ptr("notify")
	}
	if cxsettingspkg.CxSettings.UpdaterAgentEnabled == nil {
		cxsettingspkg.CxSettings.UpdaterAgentEnabled = utilities.Ptr(true)
	}
	if cxsettingspkg.CxSettings.UpdaterAgentMode == nil {
		cxsettingspkg.CxSettings.UpdaterAgentMode = utilities.Ptr("notify")
	}
	if cxsettingspkg.CxSettings.UpdaterWindowStart == nil {
		cxsettingspkg.CxSettings.UpdaterWindowStart = utilities.Ptr("00:00")
	}
	if cxsettingspkg.CxSettings.UpdaterAgentWindowStart == nil {
		cxsettingspkg.CxSettings.UpdaterAgentWindowStart = utilities.Ptr("00:00")
	}

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving serverUrl settings", err.Error())
		return
	}
	user := userFromGinContext(c)
	serverruntime.UpdateServerStatus()
	// Record activity
	_ = activitypkg.RecordUserActivity(user.ID, "System initialized", "", c.ClientIP(), c.Request.UserAgent())

	recommendation := ""
	if strings.HasPrefix(strings.ToLower(initData.ServerURL), "https://") {
		recommendation = "http"
	} else if strings.HasPrefix(strings.ToLower(initData.ServerURL), "http://") {
		recommendation = "https"
	}

	restresponse.RestSuccess(c, gin.H{
		"recommendation": recommendation,
	})
}

func finalizeInitialization(c *gin.Context) {
	var body struct {
		Disable string `json:"disable"` // "http", "https", or "none"
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	changed := false
	switch body.Disable {
	case "http":
		cxsettingspkg.CxSettings.HTTPEnabled = utilities.Ptr(false)
		changed = true
	case "https":
		cxsettingspkg.CxSettings.HTTPSEnabled = utilities.Ptr(false)
		changed = true
	}

	if changed {
		if err := cxsettingspkg.SetCxSettings(); err != nil {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
			return
		}
		// Restart listeners in background
		go func() {
			time.Sleep(500 * time.Millisecond)
			restartNetworkListenersFromSettings()
		}()
	}

	// Now logout the user
	user := userFromGinContext(c)
	if user != nil {
		RevokeAuthKey(user.AuthKey)
	}
	c.SetCookie("cxtoken", "", -1, "/", "", false, true)

	restresponse.RestSuccessNoContent(c)
}
