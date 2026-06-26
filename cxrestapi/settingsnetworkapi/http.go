package settingsnetworkapi

import (
	cxsettingspkg "chronix/internal/cxsettings"
	serverruntime "chronix/internal/serverruntime"
	"fmt"
	"strconv"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func getHTTPSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	enabled := true
	if s.HTTPEnabled != nil {
		enabled = *s.HTTPEnabled
	}
	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		enabled = true
	}
	port := fmt.Sprintf("%d", defaults.DefaultHTTPPort)
	if s.HTTPPort != nil && *s.HTTPPort > 0 {
		port = fmt.Sprintf("%d", *s.HTTPPort)
	}
	restresponse.RestSuccess(c, gin.H{
		"enabled": enabled,
		"port":    port,
	})
}

func putHTTPSettings(c *gin.Context) {
	var body struct {
		Enabled bool   `json:"enabled"`
		Port    string `json:"port"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	p := strings.TrimSpace(body.Port)
	if p == "" {
		p = fmt.Sprintf("%d", defaults.DefaultHTTPPort)
	}
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 || n >= 65536 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid HTTP port")
		return
	}

	s := cxsettingspkg.CxSettings
	httpsEnabled := true
	if s.HTTPSEnabled != nil {
		httpsEnabled = *s.HTTPSEnabled
	}
	httpsPort := defaults.DefaultHTTPSPort
	if s.HTTPSPort != nil && *s.HTTPSPort > 0 {
		httpsPort = int(*s.HTTPSPort)
	}
	agentEnabled := true
	if s.AgentEnabled != nil {
		agentEnabled = *s.AgentEnabled
	}
	agentPort := defaults.DefaultAgentPort
	if s.AgentPort != nil && *s.AgentPort > 0 {
		agentPort = int(*s.AgentPort)
	}

	if body.Enabled {
		if httpsEnabled && n == httpsPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "HTTP and HTTPS ports cannot be the same when both are enabled")
			return
		}
		if agentEnabled && n == agentPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "HTTP and Agent ports cannot be the same when both are enabled")
			return
		}
	}

	cxsettingspkg.CxSettings.HTTPEnabled = utilities.Ptr(body.Enabled)
	cxsettingspkg.CxSettings.HTTPPort = utilities.Ptr(int64(n))
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}
