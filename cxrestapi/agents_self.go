package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/agentmux"
	"chronix/internal/db"
	eventspkg "chronix/internal/events"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/datatypes"
	"gorm.io/gen/field"
)

func agentSelfRouter(app *gin.Engine) {
	_, agentPort, err := net.SplitHostPort(agentListeningAddress)
	if err != nil || agentPort == "" {
		app.POST("/agent/self/repair", AgentAuthMiddleware(), agentSelfRepair)
		app.POST("/agent/self/unregister", AgentAuthMiddleware(), agentSelfUnregister)
		return
	}
	app.POST("/agent/self/repair", allowOnlyAgentPort(agentPort), AgentAuthMiddleware(), agentSelfRepair)
	app.POST("/agent/self/unregister", allowOnlyAgentPort(agentPort), AgentAuthMiddleware(), agentSelfUnregister)
}

func agentClaims(c *gin.Context) (jwt.MapClaims, bool) {
	claimsVal, _ := c.Get("agent_claims")
	claims, ok := claimsVal.(jwt.MapClaims)
	if !ok {
		restresponse.RestErrorRespond(c, restresponse.Internal, "invalid agent claims")
		return nil, false
	}
	return claims, true
}

func agentSelfRepair(c *gin.Context) {
	claims, ok := agentClaims(c)
	if !ok {
		return
	}

	var body struct {
		Metadata map[string]any `json:"metadata"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid json")
		return
	}

	id, _ := claims["uuid"].(string)
	name, _ := claims["name"].(string)
	version, _ := claims["version"].(string)
	runningUser, _ := claims["runningUser"].(string)
	osVal, _ := claims["os"].(string)
	archVal, _ := claims["arch"].(string)
	osVersionVal, _ := claims["osVersion"].(string)
	osTypeVal, _ := claims["osType"].(string)

	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	version = strings.TrimSpace(version)
	runningUser = strings.TrimSpace(runningUser)
	osVal = strings.TrimSpace(osVal)
	archVal = strings.TrimSpace(archVal)
	osVersionVal = strings.TrimSpace(osVersionVal)
	osTypeVal = strings.TrimSpace(osTypeVal)

	if id == "" || name == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid claims: name/uuid required")
		return
	}

	now := time.Now().UTC()
	var mdPtr *datatypes.JSONMap
	if body.Metadata != nil {
		md := datatypes.JSONMap(body.Metadata)
		mdPtr = &md
	}

	assignments := []field.AssignExpr{
		db.Agent.Name.Value(name),
		db.Agent.Status.Value("active"),
		db.Agent.Suspended.Value(false),
		db.Agent.LastSeenAt.Value(now),
		db.Agent.Version.Value(version),
		db.Agent.RunningUser.Value(runningUser),
		db.Agent.Os.Value(osVal),
		db.Agent.Arch.Value(archVal),
		db.Agent.OsVersion.Value(osVersionVal),
		db.Agent.OsType.Value(osTypeVal),
	}
	if ip := strings.TrimSpace(c.ClientIP()); ip != "" {
		assignments = append(assignments, db.Agent.LastSeenIP.Value(ip))
	}
	if mdPtr != nil {
		assignments = append(assignments, db.Agent.MetadataJSON.Value(mdPtr))
	}

	if _, err := db.Agent.Where(db.Agent.UUID.Eq(id)).UpdateSimple(assignments...); err != nil {
		slog.Error("agent self repair: update", "component", "agents", "op", "self-repair", "id", id, "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
		return
	}

	_ = activitypkg.RecordUserActivity(0, "Agent repaired registration", "Agent: "+name+" ("+id+")", c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func agentSelfUnregister(c *gin.Context) {
	claims, ok := agentClaims(c)
	if !ok {
		return
	}

	id, _ := claims["uuid"].(string)
	name, _ := claims["name"].(string)
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	if id == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid claims: uuid required")
		return
	}

	now := time.Now().UTC()
	if _, err := db.Agent.Where(db.Agent.UUID.Eq(id)).UpdateSimple(
		db.Agent.Status.Value("unregistered"),
		db.Agent.Suspended.Value(true),
		db.Agent.LastSeenAt.Value(now),
	); err != nil {
		slog.Error("agent self unregister: update", "component", "agents", "op", "self-unregister", "id", id, "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
		return
	}

	wasOnline := agentmux.DefaultManager.Get(id) != nil
	agentmux.DefaultManager.Unregister(id)
	_ = eventspkg.BroadcastAdminEvent(eventspkg.SSEEventAgentDeleted, gin.H{"uuid": id, "wasOnline": wasOnline, "self": true, "status": "unregistered"})
	_ = activitypkg.RecordUserActivity(0, "Agent unregistered itself", "Agent: "+name+" ("+id+")", c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}
