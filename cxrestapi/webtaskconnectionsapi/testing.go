package webtaskconnectionsapi

import (
	"chronix/cxrestapi/apiutil"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/internal/webtaskrun"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func testWebtaskConnection(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	conn, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "connection not found")
		return
	}
	if conn.Suspended != nil && *conn.Suspended {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Connection is suspended")
		return
	}

	var p webtaskConnPayload
	if err := c.ShouldBindJSON(&p); err == nil {
		if p.BaseURL != nil {
			conn.BaseURL = p.BaseURL
		}
		if p.AgentUUID != nil {
			conn.AgentUUID = p.AgentUUID
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	ok, msg := runTestWebtaskConnection(ctx, conn.BaseURL, conn.AgentUUID, conn.AuthType, conn.AuthConfig)
	status := "ok"
	if !ok {
		status = "error"
	}
	updateWebtaskStatus(id, status, msg)
	restresponse.RestSuccess(c, gin.H{"ok": ok, "status": status, "message": msg})
}

func testWebtaskConnectionFromDraft(c *gin.Context) {
	var p webtaskConnPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var authConfig *datatypes.JSONMap
	if p.AuthConfig != nil {
		ac := datatypes.JSONMap(p.AuthConfig)
		authConfig = &ac
	}

	at := ""
	if p.AuthType != nil {
		at = *p.AuthType
	}
	ok, msg := runTestWebtaskConnection(ctx, p.BaseURL, p.AgentUUID, at, authConfig)
	status := "ok"
	if !ok {
		status = "error"
	}
	restresponse.RestSuccess(c, gin.H{"ok": ok, "status": status, "message": msg})
}

func runTestWebtaskConnection(ctx context.Context, baseURL *string, agentUUID *string, authType string, authConfig *datatypes.JSONMap) (bool, string) {
	if baseURL == nil || *baseURL == "" {
		return true, "Connection metadata is valid. No base URL to probe."
	}

	headers := make(map[string]string)
	at := strings.ToLower(authType)
	if at != "none" && authConfig != nil {
		config := *authConfig
		switch at {
		case "basic":
			user, _ := config["username"].(string)
			pass, _ := config["password"].(string)
			if pass != "" && pass != "<redacted>" {
				pass, _ = secret.Decrypt(pass)
			}
			headers["Authorization"] = "Basic " + base64.StdEncoding.EncodeToString([]byte(user+":"+pass))
		case "bearer":
			token, _ := config["token"].(string)
			if token != "" && token != "<redacted>" {
				token, _ = secret.Decrypt(token)
			}
			headers["Authorization"] = "Bearer " + token
		case "header":
			name, _ := config["header_name"].(string)
			value, _ := config["header_value"].(string)
			if value != "" && value != "<redacted>" {
				value, _ = secret.Decrypt(value)
			}
			if name != "" {
				headers[name] = value
			}
		}
	}

	var runner webtaskrun.WebTaskRunner
	if agentUUID != nil && *agentUUID != "" {
		runner = &webtaskrun.AgentRunner{AgentID: *agentUUID}
	} else {
		runner = webtaskrun.NewLocalRunner()
	}

	res, err := runner.Execute(ctx, "GET", *baseURL, headers, nil, 5*time.Second)
	if err != nil {
		return false, err.Error()
	}
	return true, fmt.Sprintf("HTTP %d", res.StatusCode)
}

func updateWebtaskStatus(id int64, status string, lastErr string) {
	now := time.Now()
	_, _ = db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).Updates(&models.WebtaskConnection{
		LastStatus:    &status,
		LastError:     &lastErr,
		LastCheckedAt: &now,
	})
}
