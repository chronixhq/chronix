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

func getHTTPSSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	mode := strings.TrimSpace(utilities.PtrVal(s.HTTPSMode))
	if mode == "" {
		mode = "selfsigned"
	}
	port := fmt.Sprintf("%d", defaults.DefaultHTTPSPort)
	if s.HTTPSPort != nil && *s.HTTPSPort > 0 {
		port = fmt.Sprintf("%d", *s.HTTPSPort)
	}
	enabled := true
	if s.HTTPSEnabled != nil {
		enabled = *s.HTTPSEnabled
	}
	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		enabled = true
	}
	certName := ""
	keyName := ""
	var certInfo gin.H
	if utilities.PtrVal(s.HTTPSCertPem) != "" {
		certName = "uploaded"
		if info, err := SummarizeCert(utilities.PtrVal(s.HTTPSCertPem)); err == nil {
			certInfo = gin.H{
				"subject":  info.Subject,
				"issuer":   info.Issuer,
				"validity": fmt.Sprintf("%s - %s", info.NotBefore, info.NotAfter),
			}
		}
	}
	if utilities.PtrVal(s.HTTPSKeyPem) != "" {
		keyName = "uploaded"
	}
	restresponse.RestSuccess(c, gin.H{
		"mode":         mode,
		"port":         port,
		"enabled":      enabled,
		"certFileName": certName,
		"keyFileName":  keyName,
		"certInfo":     certInfo,
	})
}

func putHTTPSSettings(c *gin.Context) {
	var body struct {
		Mode         string `json:"mode"`
		Port         string `json:"port"`
		Enabled      *bool  `json:"enabled"`
		CertFileName string `json:"certFileName"`
		KeyFileName  string `json:"keyFileName"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	mode := strings.TrimSpace(body.Mode)
	if mode == "" {
		mode = "selfsigned"
	}
	cxsettingspkg.CxSettings.HTTPSMode = utilities.Ptr(mode)

	p := strings.TrimSpace(body.Port)
	if p == "" {
		p = fmt.Sprintf("%d", defaults.DefaultHTTPSPort)
	}
	n, err := strconv.Atoi(p)
	if err != nil || n <= 0 || n >= 65536 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid HTTPS port")
		return
	}

	httpEnabled := false
	if cxsettingspkg.CxSettings.HTTPEnabled != nil {
		httpEnabled = *cxsettingspkg.CxSettings.HTTPEnabled
	}
	httpPort := defaults.DefaultHTTPPort
	if cxsettingspkg.CxSettings.HTTPPort != nil && *cxsettingspkg.CxSettings.HTTPPort > 0 {
		httpPort = int(*cxsettingspkg.CxSettings.HTTPPort)
	}
	agentEnabled := true
	if cxsettingspkg.CxSettings.AgentEnabled != nil {
		agentEnabled = *cxsettingspkg.CxSettings.AgentEnabled
	}
	agentPort := defaults.DefaultAgentPort
	if cxsettingspkg.CxSettings.AgentPort != nil && *cxsettingspkg.CxSettings.AgentPort > 0 {
		agentPort = int(*cxsettingspkg.CxSettings.AgentPort)
	}

	httpsEnabled := true
	if body.Enabled != nil {
		httpsEnabled = *body.Enabled
	} else if cxsettingspkg.CxSettings.HTTPSEnabled != nil {
		httpsEnabled = *cxsettingspkg.CxSettings.HTTPSEnabled
	}

	if httpsEnabled {
		if httpEnabled && n == httpPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "HTTP and HTTPS ports cannot be the same when both are enabled")
			return
		}
		if agentEnabled && n == agentPort {
			restresponse.RestErrorRespond(c, restresponse.BadRequest, "HTTPS and Agent ports cannot be the same when both are enabled")
			return
		}
	}

	cxsettingspkg.CxSettings.HTTPSPort = utilities.Ptr(int64(n))
	if body.Enabled != nil {
		cxsettingspkg.CxSettings.HTTPSEnabled = utilities.Ptr(*body.Enabled)
	}
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}
