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

func getAgentSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	enabled := true
	if s.AgentEnabled != nil {
		enabled = *s.AgentEnabled
	}
	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		enabled = true
	}
	port := fmt.Sprintf("%d", defaults.DefaultAgentPort)
	if s.AgentPort != nil && *s.AgentPort > 0 {
		port = fmt.Sprintf("%d", *s.AgentPort)
	}
	restresponse.RestSuccess(c, gin.H{
		"enabled": enabled,
		"port":    port,
	})
}

func putAgentSettings(c *gin.Context) {
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
		p = fmt.Sprintf("%d", defaults.DefaultAgentPort)
	}
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 || n >= 65536 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid Agent port")
		return
	}

	s := cxsettingspkg.CxSettings
	httpEnabled := false
	if s.HTTPEnabled != nil {
		httpEnabled = *s.HTTPEnabled
	}
	httpPort := defaults.DefaultHTTPPort
	if s.HTTPPort != nil && *s.HTTPPort > 0 {
		httpPort = int(*s.HTTPPort)
	}
	httpsEnabled := true
	if s.HTTPSEnabled != nil {
		httpsEnabled = *s.HTTPSEnabled
	}
	httpsPort := defaults.DefaultHTTPSPort
	if s.HTTPSPort != nil && *s.HTTPSPort > 0 {
		httpsPort = int(*s.HTTPSPort)
	}

	if body.Enabled {
		if httpEnabled && n == httpPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "Agent and HTTP ports cannot be the same when both are enabled")
			return
		}
		if httpsEnabled && n == httpsPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "Agent and HTTPS ports cannot be the same when both are enabled")
			return
		}
	}

	cxsettingspkg.CxSettings.AgentEnabled = utilities.Ptr(body.Enabled)
	cxsettingspkg.CxSettings.AgentPort = utilities.Ptr(int64(n))
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}
