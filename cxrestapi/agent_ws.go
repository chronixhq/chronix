package cxrestapi

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	activitypkg "chronix/internal/activity"
	"chronix/internal/agentmux"
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	notifypkg "chronix/internal/notify"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

// agentWSRouter mounts the WebSocket connect endpoint for Agents.
func agentWSRouter(app *gin.Engine) {
	// Wire agent online/offline notifications via agentmux callbacks (Always On)
	agentmux.OnConnect = func(id, name string) {
		details := fmt.Sprintf("Agent: %s (%s)", name, id)
		_ = activitypkg.RecordUserActivity(0, "Agent connected", details, "", "")
	}
	agentmux.OnDisconnect = func(id, name string) {
		details := fmt.Sprintf("Agent: %s (%s)", name, id)
		_ = activitypkg.RecordUserActivity(0, "Agent disconnected", details, "", "")

		cxs := cxsettingspkg.GetCxSettings()
		if cxs.AlertOnAgentLost != nil && *cxs.AlertOnAgentLost {
			data := map[string]any{"agent_id": id, "agent_name": name}
			notifypkg.TryCreateNotification(notifypkg.CategorySystem, notifypkg.SeverityError, "Agent disconnected", nil, &data)
		}
	}
	// Restrict this route to only accept traffic that lands on the Agent listener's port.
	_, agentPort, err := net.SplitHostPort(agentListeningAddress)
	if err != nil || agentPort == "" {
		// Fallback: do not apply guard if parsing fails; still register endpoint.
		app.GET("/agent/connect", AgentAuthMiddleware(), agentWSConnect)
		return
	}
	app.GET("/agent/connect", allowOnlyAgentPort(agentPort), AgentAuthMiddleware(), agentWSConnect)
}

// allowOnlyAgentPort ensures the handler only serves requests for the configured agent port.
func allowOnlyAgentPort(port string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Enforce TLS at the handler level for Agent WS
		if c.Request.TLS == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		host := c.Request.Host
		if !strings.Contains(host, ":") {
			// Default TLS port assumption if no explicit port in Host header
			if c.Request.TLS != nil {
				if port != "443" {
					c.AbortWithStatus(http.StatusNotFound)
					return
				}
			} else {
				// Non-TLS default 80, which we don't expect here
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		} else {
			_, reqPort, _ := net.SplitHostPort(host)
			if reqPort != port {
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}
		c.Next()
	}
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// For MVP allow any origin; tighten later if needed
	CheckOrigin: func(_ *http.Request) bool { return true },
}

func AgentAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authz := c.GetHeader("Authorization")
		if !strings.HasPrefix(strings.ToLower(authz), "bearer ") {
			restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "missing bearer token")
			c.Abort()
			return
		}
		tokenStr := strings.TrimSpace(authz[len("Bearer "):])
		claims := jwt.MapClaims{}
		keyFunc := func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodEdDSA {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			mc, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return nil, jwt.ErrTokenInvalidClaims
			}
			uuidVal, _ := mc["uuid"].(string)
			uuidVal = strings.TrimSpace(uuidVal)
			if uuidVal == "" {
				return nil, jwt.ErrTokenInvalidClaims
			}
			ag, err := db.Agent.Where(db.Agent.UUID.Eq(uuidVal)).Take()
			if err != nil {
				slog.Error("agent connect: db lookup failed", "uuid", uuidVal, "error", err, "component", "agent-ws", "op", "auth")
				return nil, jwt.ErrTokenInvalidSubject
			}
			if ag == nil {
				slog.Warn("agent connect: agent not found", "uuid", uuidVal, "component", "agent-ws", "op", "auth")
				return nil, jwt.ErrTokenInvalidSubject
			}
			if strings.ToLower(ag.Status) != "active" {
				slog.Warn("agent connect: agent not active", "uuid", uuidVal, "status", ag.Status, "component", "agent-ws", "op", "auth")
				return nil, jwt.ErrTokenInvalidSubject
			}
			if ag.Suspended != nil && *ag.Suspended {
				slog.Warn("agent connect: agent suspended", "uuid", uuidVal, "component", "agent-ws", "op", "auth")
				return nil, jwt.ErrTokenInvalidSubject
			}
			if ag.PublicKey == nil || *ag.PublicKey == "" {
				slog.Error("agent connect: agent missing public key", "uuid", uuidVal, "component", "agent-ws", "op", "auth")
				return nil, jwt.ErrTokenInvalidSubject
			}
			pubBytes, err := base64.StdEncoding.DecodeString(*ag.PublicKey)
			if err != nil {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			if len(pubBytes) != ed25519.PublicKeySize {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return ed25519.PublicKey(pubBytes), nil
		}
		token, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc, jwt.WithValidMethods([]string{"EdDSA"}))
		if err != nil || !token.Valid {
			slog.Debug("invalid agent jwt", "component", "agent-ws", "op", "auth", "error", err)
			restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "invalid token")
			c.Abort()
			return
		}
		c.Set("agent_claims", claims)
		c.Next()
	}
}

// agentWSConnect upgrades to WebSocket and registers the connection with agentmux.
func agentWSConnect(c *gin.Context) {
	claimsVal, _ := c.Get("agent_claims")
	claims, ok := claimsVal.(jwt.MapClaims)
	if !ok {
		restresponse.RestErrorRespond(c, restresponse.Internal, "invalid agent claims")
		return
	}
	name, _ := claims["name"].(string)
	id, _ := claims["uuid"].(string)
	version, _ := claims["version"].(string)
	runningUser, _ := claims["runningUser"].(string)
	osVal, _ := claims["os"].(string)
	archVal, _ := claims["arch"].(string)
	osVersionVal, _ := claims["osVersion"].(string)
	osTypeVal, _ := claims["osType"].(string)

	name = strings.TrimSpace(name)
	id = strings.TrimSpace(id)
	version = strings.TrimSpace(version)
	runningUser = strings.TrimSpace(runningUser)
	osVal = strings.TrimSpace(osVal)
	archVal = strings.TrimSpace(archVal)
	osVersionVal = strings.TrimSpace(osVersionVal)
	osTypeVal = strings.TrimSpace(osTypeVal)

	if name == "" || id == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid claims: name/uuid required")
		return
	}
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("ws upgrade failed", "component", "agent-ws", "op", "upgrade", "error", err)
		return
	}
	// Heartbeat/idle handling: update read deadline on control frames
	const pingInterval = 30 * time.Second
	const idleTimeout = pingInterval * 3 // 90s
	_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
	_ = conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(idleTimeout))
	})
	// When client pings us (expected every 30s), extend the deadline and reply with Pong
	conn.SetPingHandler(func(appData string) error {
		_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
	})
	remoteAddr, _ := net.ResolveTCPAddr("tcp", c.Request.RemoteAddr)
	agentmux.DefaultManager.Register(id, name, conn, remoteAddr)
	// Update last seen fields on connect
	ip := c.ClientIP()
	now := time.Now().UTC()
	if ip != "" {
		ipCopy := ip
		if _, err := db.Agent.Where(db.Agent.UUID.Eq(id)).UpdateSimple(
			db.Agent.LastSeenIP.Value(ipCopy),
			db.Agent.LastSeenAt.Value(now),
			db.Agent.Version.Value(version),
			db.Agent.RunningUser.Value(runningUser),
			db.Agent.Os.Value(osVal),
			db.Agent.Arch.Value(archVal),
			db.Agent.OsVersion.Value(osVersionVal),
			db.Agent.OsType.Value(osTypeVal),
		); err != nil {
			slog.Error("agent last-seen update (connect)", "component", "agent-ws", "op", "connect", "id", id, "agent", name, "error", err)
		}
	} else {
		if _, err := db.Agent.Where(db.Agent.UUID.Eq(id)).UpdateSimple(
			db.Agent.LastSeenAt.Value(now),
			db.Agent.Version.Value(version),
			db.Agent.RunningUser.Value(runningUser),
			db.Agent.Os.Value(osVal),
			db.Agent.Arch.Value(archVal),
			db.Agent.OsVersion.Value(osVersionVal),
			db.Agent.OsType.Value(osTypeVal),
		); err != nil {
			slog.Error("agent last-seen update (connect-noip)", "component", "agent-ws", "op", "connect", "id", id, "agent", name, "error", err)
		}
	}
	// Throttle updates on pongs to avoid excessive writes
	var lastDBUpdate time.Time
	conn.SetPongHandler(func(_ string) error {
		_ = conn.SetReadDeadline(time.Now().Add(idleTimeout))
		if time.Since(lastDBUpdate) >= 30*time.Second {
			lastDBUpdate = time.Now()
			go func(u string, ip string, t time.Time) {
				if ip != "" {
					ipCopy := ip
					_, err := db.Agent.Where(db.Agent.UUID.Eq(u)).UpdateSimple(db.Agent.LastSeenIP.Value(ipCopy), db.Agent.LastSeenAt.Value(t))
					if err != nil {
						slog.Error("agent last-seen update (pong)", "component", "agent-ws", "op", "pong", "id", u, "error", err)
					}
				} else {
					_, err := db.Agent.Where(db.Agent.UUID.Eq(u)).UpdateSimple(db.Agent.LastSeenAt.Value(t))
					if err != nil {
						slog.Error("agent last-seen update (pong-noip)", "component", "agent-ws", "op", "pong", "id", u, "error", err)
					}
				}
			}(id, ip, time.Now().UTC())
		}
		return nil
	})
	slog.Info("agent ws connected", "component", "agent-ws", "op", "connect", "agent", name, "id", id, "remote", c.Request.RemoteAddr)
}
