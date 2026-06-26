package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/agentmux"
	"chronix/internal/db"
	"chronix/internal/db/models"
	eventspkg "chronix/internal/events"
	"chronix/internal/updater"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func agentsUpdate(c *gin.Context) {
	requestUUID := strings.TrimSpace(c.Param("uuid"))
	if requestUUID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing uuid")
		return
	}
	if updater.AvailableVersion == nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "no update available")
		return
	}
	agConn := agentmux.DefaultManager.Get(requestUUID)
	if agConn == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "agent not online")
		return
	}

	user := userFromGinContext(c)
	version := updater.AvailableVersion.Agent.Version
	slog.Info("Agent update initiated via UI", "agentUuid", requestUUID, "version", version, "user", user.Email)
	_ = activitypkg.RecordUserActivity(user.ID, "Agent Update Initiated", fmt.Sprintf("Updated agent %s to version %s", requestUUID, version), c.ClientIP(), c.Request.UserAgent())

	_, _, err := agConn.Request(c.Request.Context(), "agent.update", gin.H{"version": version})
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to send update command to agent", err.Error())
		return
	}
	restresponse.RestSuccess(c, gin.H{"status": "update initiated"})
}

func agentsRestart(c *gin.Context) {
	requestUUID := strings.TrimSpace(c.Param("uuid"))
	if requestUUID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing uuid")
		return
	}

	agConn := agentmux.DefaultManager.Get(requestUUID)
	if agConn == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "agent not online")
		return
	}

	user := userFromGinContext(c)
	slog.Info("Agent restart initiated via UI", "agentUuid", requestUUID, "user", user.Email)
	_ = activitypkg.RecordUserActivity(user.ID, "Agent Restart Initiated", fmt.Sprintf("Restarted agent %s", requestUUID), c.ClientIP(), c.Request.UserAgent())

	_, _, err := agConn.Request(c.Request.Context(), "agent.restart", nil)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to send restart command to agent", err.Error())
		return
	}
	restresponse.RestSuccess(c, gin.H{"status": "restart initiated"})
}

func getAgent(c *gin.Context) {
	requestUUID := strings.TrimSpace(c.Param("uuid"))
	if requestUUID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing uuid")
		return
	}
	ag, err := db.Agent.Where(db.Agent.UUID.Eq(requestUUID)).Take()
	if err != nil || ag == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "agent not found")
		return
	}
	restresponse.RestSuccess(c, gin.H{
		"uuid":         ag.UUID,
		"name":         ag.Name,
		"status":       ag.Status,
		"suspended":    ag.Suspended,
		"version":      utilities.PtrVal(ag.Version),
		"os":           ag.Os,
		"arch":         ag.Arch,
		"osVersion":    ag.OsVersion,
		"osType":       ag.OsType,
		"runningUser":  ag.RunningUser,
		"publicKey":    ag.PublicKey,
		"approvedBy":   ag.ApprovedByUserID,
		"approvedAt":   ag.ApprovedAt,
		"lastSeenAt":   ag.LastSeenAt,
		"lastSeenIp":   ag.LastSeenIP,
		"metadataJson": ag.MetadataJSON,
	})
}

func agentsApprove(c *gin.Context) {
	reqID := c.Param("requestId")
	if strings.TrimSpace(reqID) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing requestId")
		return
	}
	req, err := db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).Take()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "request not found")
		return
	}
	now := time.Now().UTC()
	if now.After(req.ExpiresAt) || strings.ToLower(req.Status) != "pending" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "request is not pending or has expired")
		return
	}
	user := userFromGinContext(c)

	existing, _ := db.Agent.Where(db.Agent.UUID.Eq(req.UUID)).Take()
	if existing != nil && existing.UUID == req.UUID {
		_, err = db.Agent.Where(db.Agent.UUID.Eq(req.UUID)).UpdateSimple(
			db.Agent.Name.Value(req.Name),
			db.Agent.PublicKey.Value(req.PublicKey),
			db.Agent.Version.Value(utilities.PtrVal(req.Version)),
			db.Agent.Os.Value(utilities.PtrVal(req.Os)),
			db.Agent.Arch.Value(utilities.PtrVal(req.Arch)),
			db.Agent.OsVersion.Value(utilities.PtrVal(req.OsVersion)),
			db.Agent.OsType.Value(utilities.PtrVal(req.OsType)),
			db.Agent.Status.Value("active"),
			db.Agent.Suspended.Value(false),
			db.Agent.ApprovedByUserID.Value(user.ID),
			db.Agent.ApprovedAt.Value(now),
			db.Agent.MetadataJSON.Value(req.MetadataJSON),
		)
		if err != nil {
			slog.Error("approve: update agent", "component", "agents", "op", "approve", "error", err)
			restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
			return
		}
	} else {
		pk := req.PublicKey
		uid := user.ID
		nowCopy := now
		ag := &models.Agent{
			UUID:             req.UUID,
			Name:             req.Name,
			Status:           "active",
			Suspended:        utilities.Ptr(false),
			PublicKey:        &pk,
			Version:          req.Version,
			Os:               req.Os,
			Arch:             req.Arch,
			OsVersion:        req.OsVersion,
			OsType:           req.OsType,
			RunningUser:      req.RunningUser,
			ApprovedByUserID: &uid,
			ApprovedAt:       &nowCopy,
			MetadataJSON:     req.MetadataJSON,
		}
		if err := db.Agent.Create(ag); err != nil {
			slog.Error("approve: create agent", "component", "agents", "op", "approve", "error", err)
			restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
			return
		}
	}

	_, _ = db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).UpdateSimple(
		db.AgentRegistrationRequest.Status.Value("approved"),
		db.AgentRegistrationRequest.ApprovedByUserID.Value(user.ID),
		db.AgentRegistrationRequest.ApprovedAt.Value(now),
		db.AgentRegistrationRequest.ConsumedAt.Value(now),
	)
	_ = eventspkg.BroadcastAdminEvent(eventspkg.SSEEventAgentRegApproved, gin.H{"requestId": reqID, "uuid": req.UUID, "name": req.Name})
	_ = activitypkg.RecordUserActivity(user.ID, "Approved agent", fmt.Sprintf("Agent: %s (%s)", req.Name, req.UUID), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func agentsDeny(c *gin.Context) {
	reqID := c.Param("requestId")
	if strings.TrimSpace(reqID) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing requestId")
		return
	}
	req, err := db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).Take()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "request not found")
		return
	}
	user := userFromGinContext(c)
	now := time.Now()
	_, _ = db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).UpdateSimple(
		db.AgentRegistrationRequest.Status.Value("denied"),
		db.AgentRegistrationRequest.ApprovedByUserID.Value(user.ID),
		db.AgentRegistrationRequest.ApprovedAt.Value(now),
	)
	_ = eventspkg.BroadcastAdminEvent(eventspkg.SSEEventAgentRegDenied, gin.H{"requestId": reqID, "uuid": req.UUID, "name": req.Name})
	_ = activitypkg.RecordUserActivity(user.ID, "Denied agent", fmt.Sprintf("Agent: %s (%s)", req.Name, req.UUID), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func agentsDelete(c *gin.Context) {
	uuidStr := strings.TrimSpace(c.Param("uuid"))
	if uuidStr == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing uuid")
		return
	}
	ag, err := db.Agent.Where(db.Agent.UUID.Eq(uuidStr)).Take()
	if err != nil || ag == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "agent not found")
		return
	}

	type connRef struct {
		ID   any    `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	}
	var conns []connRef
	dbConns, _ := db.DbConnection.Where(db.DbConnection.AgentUUID.Eq(uuidStr)).Select(db.DbConnection.ID, db.DbConnection.Name).Find()
	for _, it := range dbConns {
		conns = append(conns, connRef{ID: it.ID, Name: it.Name, Type: "database"})
	}
	shConns, _ := db.ShellConnection.Where(db.ShellConnection.AgentUUID.Eq(uuidStr)).Select(db.ShellConnection.ID, db.ShellConnection.Name).Find()
	for _, it := range shConns {
		conns = append(conns, connRef{ID: it.ID, Name: it.Name, Type: "shell"})
	}
	wtConns, _ := db.WebtaskConnection.Where(db.WebtaskConnection.AgentUUID.Eq(uuidStr)).Select(db.WebtaskConnection.ID, db.WebtaskConnection.Name).Find()
	for _, it := range wtConns {
		conns = append(conns, connRef{ID: it.ID, Name: it.Name, Type: "webtask"})
	}

	inUseCount := len(conns)
	preview := strings.EqualFold(c.Query("preview"), "true")
	removeMappings := strings.EqualFold(c.Query("removeMappings"), "true")
	if preview {
		restresponse.RestSuccess(c, gin.H{"agentUuid": uuidStr, "inUseCount": inUseCount, "connections": conns})
		return
	}
	if inUseCount > 0 && !removeMappings {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "agent is in use by connections", gin.H{"connections": conns})
		return
	}

	var removed int64
	if err := db.Q.Transaction(func(tx *db.Query) error {
		if removeMappings {
			info, err := tx.DbConnection.Where(tx.DbConnection.AgentUUID.Eq(uuidStr)).Update(tx.DbConnection.AgentUUID, nil)
			if err != nil {
				return err
			}
			removed += info.RowsAffected
			info, err = tx.ShellConnection.Where(tx.ShellConnection.AgentUUID.Eq(uuidStr)).Update(tx.ShellConnection.AgentUUID, nil)
			if err != nil {
				return err
			}
			removed += info.RowsAffected
			info, err = tx.WebtaskConnection.Where(tx.WebtaskConnection.AgentUUID.Eq(uuidStr)).Update(tx.WebtaskConnection.AgentUUID, nil)
			if err != nil {
				return err
			}
			removed += info.RowsAffected
		}
		if _, err := tx.Agent.Where(tx.Agent.UUID.Eq(uuidStr)).Delete(); err != nil {
			return err
		}
		return nil
	}); err != nil {
		slog.Error("agent delete: tx", "component", "agents", "op", "delete", "id", uuidStr, "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
		return
	}

	wasOnline := false
	acked := false
	if conn := agentmux.DefaultManager.Get(uuidStr); conn != nil {
		wasOnline = true
		if ok, _, err := agentmux.DefaultManager.NotifyAgentDeleted(uuidStr, "deleted via admin", 3*time.Second); err == nil {
			acked = ok
		} else {
			slog.Warn("agent delete notify: ack timeout or error", "component", "agents", "op", "delete", "id", uuidStr, "error", err)
		}
		agentmux.DefaultManager.Unregister(uuidStr)
	}
	_ = eventspkg.BroadcastAdminEvent(eventspkg.SSEEventAgentDeleted, gin.H{"uuid": uuidStr, "wasOnline": wasOnline, "acked": acked})
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted agent", fmt.Sprintf("Agent: %s (%s)", ag.Name, uuidStr), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"deleted": true, "removedMappings": removed, "agentUuid": uuidStr, "acked": acked, "wasOnline": wasOnline})
}
