package cxrestapi

import (
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	"chronix/internal/db/models"
	eventspkg "chronix/internal/events"
	"encoding/base64"
	"log/slog"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

func agentsRegisterBegin(c *gin.Context) {
	if !isAgentPort(c) {
		restresponse.RestErrorRespond(c, restresponse.PermissionDenied, "Agent registration only allowed on the Agent Port")
		return
	}
	agentEnabled := true
	if cxsettingspkg.CxSettings.AgentEnabled != nil {
		agentEnabled = *cxsettingspkg.CxSettings.AgentEnabled
	}
	if !agentEnabled {
		restresponse.RestErrorRespond(c, restresponse.PermissionDenied, "agent connections are disabled")
		return
	}

	ip := c.ClientIP()
	if !rl.allow(ip) {
		restresponse.RestErrorRespond(c, restresponse.TooManyRequests, "too many requests")
		return
	}

	var body struct {
		Name      string         `json:"name"`
		UUID      string         `json:"uuid"`
		Version   string         `json:"version"`
		PublicKey string         `json:"publicKey"`
		Metadata  map[string]any `json:"metadata"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid json")
		return
	}

	name := strings.TrimSpace(body.Name)
	agentID := strings.TrimSpace(body.UUID)
	pk := strings.TrimSpace(body.PublicKey)
	version := strings.TrimSpace(body.Version)
	if name == "" || agentID == "" || pk == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "name, uuid, and publicKey are required")
		return
	}
	if _, err := base64.StdEncoding.DecodeString(pk); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid publicKey encoding")
		return
	}

	requestID := uuid.NewString()
	now := time.Now().UTC()
	expires := now.Add(5 * time.Minute)
	var mdPtr *datatypes.JSONMap
	var runningUser *string
	var os, arch, osVersion, osType *string
	if body.Metadata != nil {
		m := datatypes.JSONMap(body.Metadata)
		mdPtr = &m
		if u, ok := body.Metadata["user"].(string); ok {
			runningUser = &u
		}
		if v, ok := body.Metadata["os"].(string); ok {
			os = &v
		}
		if v, ok := body.Metadata["arch"].(string); ok {
			arch = &v
		}
		if v, ok := body.Metadata["os_version"].(string); ok {
			osVersion = &v
		}
		if v, ok := body.Metadata["os_type"].(string); ok {
			osType = &v
		}
	}

	ipCopy := ip
	req := &models.AgentRegistrationRequest{
		RequestID:    requestID,
		UUID:         agentID,
		Name:         name,
		PublicKey:    pk,
		Version:      &version,
		Os:           os,
		Arch:         arch,
		OsVersion:    osVersion,
		OsType:       osType,
		RunningUser:  runningUser,
		IP:           &ipCopy,
		MetadataJSON: mdPtr,
		Status:       "pending",
		CreatedAt:    &now,
		ExpiresAt:    expires,
	}
	if err := db.AgentRegistrationRequest.Create(req); err != nil {
		slog.Error("agent register-begin: db create", "component", "agents", "op", "register-begin", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "database error")
		return
	}

	_ = eventspkg.BroadcastAdminEvent(eventspkg.SSEEventAgentRegistration, gin.H{
		"requestId": requestID,
		"uuid":      agentID,
		"name":      name,
		"ip":        ip,
		"metadata":  body.Metadata,
		"expiresAt": expires,
	})
	restresponse.RestSuccess(c, gin.H{"requestId": requestID, "pollAfterMs": 1000})
}

func agentsRegisterStatus(c *gin.Context) {
	if !isAgentPort(c) {
		restresponse.RestErrorRespond(c, restresponse.PermissionDenied, "Agent registration status only allowed on the Agent Port")
		return
	}
	reqID := c.Param("requestId")
	if strings.TrimSpace(reqID) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "missing requestId")
		return
	}
	req, err := db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).Take()
	if err != nil {
		restresponse.RestSuccess(c, gin.H{"status": "expired"})
		return
	}
	now := time.Now().UTC()
	status := strings.ToLower(req.Status)
	if status == "pending" && now.After(req.ExpiresAt) {
		_, _ = db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.RequestID.Eq(reqID)).UpdateSimple(db.AgentRegistrationRequest.Status.Value("expired"))
		status = "expired"
	}
	restresponse.RestSuccess(c, gin.H{"status": status})
}
