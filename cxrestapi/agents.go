package cxrestapi

import (
	"chronix/internal/agentmux"
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"fmt"
	"net"
	"time"

	"net/http"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func listAgents(c *gin.Context) {
	// If manifest is nil or older than 5 minutes, refresh it
	if updater.AvailableVersion == nil || time.Since(updater.LastCheckTime) > 5*time.Minute {
		_, _, _ = updater.CheckForUpdates(buildinfo.Version)
	}

	availableAgentVersion := ""
	if updater.AvailableVersion != nil {
		availableAgentVersion = updater.AvailableVersion.Agent.Version
	}

	items, err := db.Agent.Order(db.Agent.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list agents", err.Error())
		return
	}
	resp := make([]gin.H, 0, len(items))
	// Consider an agent online if seen within the last 2 minutes
	onlineThreshold := 2 * time.Minute
	// Build a set of currently connected agent UUIDs from the live manager
	liveIDs := map[string]struct{}{}
	for _, id := range agentmux.DefaultManager.List() {
		liveIDs[id] = struct{}{}
	}
	for _, a := range items {
		// Prefer live connection info from agentmux; fallback to recent lastSeenAt
		_, isLive := liveIDs[a.UUID]
		online := isLive
		if !online && a.LastSeenAt != nil {
			if time.Since(*a.LastSeenAt) <= onlineThreshold {
				online = true
			}
		}

		version := utilities.PtrVal(a.Version)
		updateAvailable := ""
		if availableAgentVersion != "" && version != availableAgentVersion {
			updateAvailable = availableAgentVersion
		}

		resp = append(resp, gin.H{
			"uuid":            a.UUID,
			"name":            a.Name,
			"version":         version,
			"os":              a.Os,
			"arch":            a.Arch,
			"osVersion":       a.OsVersion,
			"osType":          a.OsType,
			"runningUser":     a.RunningUser,
			"updateAvailable": updateAvailable,
			"status":          a.Status,
			"suspended":       a.Suspended,
			"online":          online,
			"lastSeenAt":      a.LastSeenAt,
			"lastSeenIp":      a.LastSeenIP,
		})
	}
	restresponse.RestSuccess(c, resp)
}

// ---- In-memory rate limiter for register-begin ----
// very simple sliding window per IP for MVP
var rl = newRateLimiter(5, time.Minute) // 5 per minute per IP

type rateLimiter struct {
	limit int
	win   time.Duration
	data  map[string][]time.Time
}

func newRateLimiter(limit int, win time.Duration) *rateLimiter {
	return &rateLimiter{limit: limit, win: win, data: make(map[string][]time.Time)}
}

func (r *rateLimiter) allow(ip string) bool {
	now := time.Now()
	list := r.data[ip]
	// drop old
	i := 0
	for ; i < len(list); i++ {
		if now.Sub(list[i]) <= r.win {
			break
		}
	}
	if i > 0 {
		list = list[i:]
	}
	if len(list) >= r.limit {
		r.data[ip] = list
		return false
	}
	list = append(list, now)
	r.data[ip] = list
	return true
}

func isAgentPort(c *gin.Context) bool {
	localAddr, ok := c.Request.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok {
		return true
	}
	_, portStr, err := net.SplitHostPort(localAddr.String())
	if err != nil {
		return true
	}

	expectedPort := DefaultAgentPort
	if cxsettingspkg.CxSettings.AgentPort != nil && *cxsettingspkg.CxSettings.AgentPort > 0 {
		expectedPort = int(*cxsettingspkg.CxSettings.AgentPort)
	}
	if CLIHTTPConfig.ForceAgentPort > 0 {
		expectedPort = int(CLIHTTPConfig.ForceAgentPort)
	}

	return portStr == fmt.Sprintf("%d", expectedPort)
}

func agentsAnonymousRouter(app *gin.Engine) {
	app.POST("/agent/register", agentsRegisterBegin)
	app.GET("/agent/register/status/:requestId", agentsRegisterStatus)
}

func agentsAdminRouter(app *gin.Engine) {
	app.GET("/agents", listAgents)
	app.GET("/agents/:uuid", getAgent)
	app.DELETE("/agents/:uuid", adminFunc(agentsDelete))
	app.POST("/agents/:uuid/update", adminFunc(agentsUpdate))
	app.POST("/agents/:uuid/restart", adminFunc(agentsRestart))
	app.POST("/agents/requests/:requestId/approve", adminFunc(agentsApprove))
	app.POST("/agents/requests/:requestId/deny", adminFunc(agentsDeny))
}
