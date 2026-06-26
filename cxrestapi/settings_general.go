package cxrestapi

import (
	"chronix/cxrestapi/settingsnetworkapi"
	cxsettingspkg "chronix/internal/cxsettings"
	eventspkg "chronix/internal/events"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/updater"
	"log/slog"
	"os"
	"strings"
	"time"

	app_settings "github.com/dan-sherwin/go-app-settings"
	appsettings_models "github.com/dan-sherwin/go-app-settings/db/models"
	restapi "github.com/dan-sherwin/go-rest-api-server"
	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func setServerURL(c *gin.Context) {
	var data utilities.StrMap
	if err := c.BindJSON(&data); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	serverURL, ok := data["serverUrl"]
	if !ok {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Missing serverUrl")
		return
	}
	cxsettingspkg.CxSettings.ServerURL = &serverURL
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func getServerURL(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	restresponse.RestSuccess(c, gin.H{"serverUrl": utilities.PtrVal(s.ServerURL)})
}

func putServerURL(c *gin.Context) {
	var body struct {
		ServerURL string `json:"serverUrl"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.ServerURL = utilities.Ptr(strings.TrimSpace(body.ServerURL))
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func getGlobalAlertSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	resp := gin.H{
		"systemAlertEmails": utilities.PtrVal(s.SystemAlertEmails),
		"systemAlertPhones": utilities.PtrVal(s.SystemAlertPhones),
		"alertOnAgentLost":  utilities.PtrVal(s.AlertOnAgentLost),
	}
	restresponse.RestSuccess(c, resp)
}

func putGlobalAlertSettings(c *gin.Context) {
	var body struct {
		SystemAlertEmails string `json:"systemAlertEmails"`
		SystemAlertPhones string `json:"systemAlertPhones"`
		AlertOnAgentLost  bool   `json:"alertOnAgentLost"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.SystemAlertEmails = utilities.Ptr(strings.TrimSpace(body.SystemAlertEmails))
	cxsettingspkg.CxSettings.SystemAlertPhones = utilities.Ptr(strings.TrimSpace(body.SystemAlertPhones))
	cxsettingspkg.CxSettings.AlertOnAgentLost = utilities.Ptr(body.AlertOnAgentLost)

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func getUpdaterSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	resp := gin.H{
		"enabled":          utilities.PtrVal(s.UpdaterEnabled),
		"mode":             utilities.PtrVal(s.UpdaterMode),
		"windowStart":      utilities.PtrVal(s.UpdaterWindowStart),
		"agentEnabled":     utilities.PtrVal(s.UpdaterAgentEnabled),
		"agentMode":        utilities.PtrVal(s.UpdaterAgentMode),
		"agentWindowStart": utilities.PtrVal(s.UpdaterAgentWindowStart),
	}
	restresponse.RestSuccess(c, resp)
}

func putUpdaterAppSettings(c *gin.Context) {
	var body struct {
		Enabled     bool   `json:"enabled"`
		Mode        string `json:"mode"`
		WindowStart string `json:"windowStart"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.UpdaterEnabled = utilities.Ptr(body.Enabled)
	cxsettingspkg.CxSettings.UpdaterMode = utilities.Ptr(strings.TrimSpace(body.Mode))
	cxsettingspkg.CxSettings.UpdaterWindowStart = utilities.Ptr(strings.TrimSpace(body.WindowStart))

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	cxsettingspkg.SyncUpdaterSettings()

	restresponse.RestSuccessNoContent(c)
}

func putUpdaterAgentSettings(c *gin.Context) {
	var body struct {
		Enabled     bool   `json:"enabled"`
		Mode        string `json:"mode"`
		WindowStart string `json:"windowStart"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.UpdaterAgentEnabled = utilities.Ptr(body.Enabled)
	cxsettingspkg.CxSettings.UpdaterAgentMode = utilities.Ptr(strings.TrimSpace(body.Mode))
	cxsettingspkg.CxSettings.UpdaterAgentWindowStart = utilities.Ptr(strings.TrimSpace(body.WindowStart))

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	cxsettingspkg.SyncUpdaterSettings()

	restresponse.RestSuccessNoContent(c)
}

func getSettingsSummary(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	httpsEnabled := true
	if s.HTTPSEnabled != nil {
		httpsEnabled = *s.HTTPSEnabled
	}
	httpEnabled := false
	if s.HTTPEnabled != nil {
		httpEnabled = *s.HTTPEnabled
	}
	agentEnabled := true
	if s.AgentEnabled != nil {
		agentEnabled = *s.AgentEnabled
	}

	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		httpsEnabled = true
		httpEnabled = true
		agentEnabled = true
	}

	httpPort := DefaultHTTPPort
	if s.HTTPPort != nil && *s.HTTPPort > 0 {
		httpPort = int(*s.HTTPPort)
	}
	httpsPort := DefaultHTTPSPort
	if s.HTTPSPort != nil && *s.HTTPSPort > 0 {
		httpsPort = int(*s.HTTPSPort)
	}
	mode := utilities.PtrVal(s.HTTPSMode)
	if mode == "" {
		mode = "selfsigned"
	}
	email := gin.H{
		"smtpHost": utilities.PtrVal(s.SMTPHost),
		"smtpPort": func() int {
			if s.SMTPPort != nil {
				return int(*s.SMTPPort)
			}
			return 587
		}(),
		"secure":     utilities.PtrVal(s.SMTPSecure),
		"fromName":   utilities.PtrVal(s.SMTPFromName),
		"fromEmail":  utilities.PtrVal(s.SMTPFromEmail),
		"configured": utilities.PtrVal(s.SMTPHost) != "" && utilities.PtrVal(s.SMTPFromEmail) != "",
	}
	var sms gin.H
	if utilities.PtrVal(s.SmsProvider) != "" {
		sms = gin.H{
			"provider":   utilities.PtrVal(s.SmsProvider),
			"fromNumber": utilities.PtrVal(s.SmsFromNumber),
			"configured": true,
		}
	} else {
		sms = gin.H{"configured": false}
	}
	var certInfo gin.H
	if utilities.PtrVal(s.HTTPSCertPem) != "" {
		if info, err := settingsnetworkapi.SummarizeCert(utilities.PtrVal(s.HTTPSCertPem)); err == nil {
			certInfo = gin.H{"subject": info.Subject, "issuer": info.Issuer, "notBefore": info.NotBefore, "notAfter": info.NotAfter}
		}
	}
	agentPort := DefaultAgentPort
	if s.AgentPort != nil && *s.AgentPort > 0 {
		agentPort = int(*s.AgentPort)
	}
	resp := gin.H{
		"serverUrl": utilities.PtrVal(s.ServerURL),
		"http":      gin.H{"enabled": httpEnabled, "port": httpPort},
		"https":     gin.H{"enabled": httpsEnabled, "port": httpsPort, "mode": mode, "hasUploadedCert": utilities.PtrVal(s.HTTPSCertPem) != "", "hasUploadedKey": utilities.PtrVal(s.HTTPSKeyPem) != "", "certInfo": certInfo},
		"agent":     gin.H{"enabled": agentEnabled, "port": agentPort},
		"email":     email,
		"sms":       sms,
	}
	restresponse.RestSuccess(c, resp)
}

func getActiveSettings(c *gin.Context) {
	var running []appsettings_models.AppSetting
	cmd := app_settings.SettingsListRunningCommand{}
	if err := cmd.GetRunningSettings(&struct{}{}, &running); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error listing active settings", err.Error())
		return
	}

	redact := func(name string, val string) string {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "password") || strings.Contains(lower, "secret") || strings.Contains(lower, "token") || strings.HasSuffix(lower, "_key") || strings.Contains(lower, "privatekey") {
			if val == "" {
				return ""
			}
			return "••••••"
		}
		return val
	}

	resp := make([]gin.H, 0, len(running))
	for _, s := range running {
		resp = append(resp, gin.H{
			"setting":     s.Key,
			"value":       redact(s.Key, s.Value),
			"description": s.Description,
		})
	}
	restresponse.RestSuccess(c, gin.H{"settings": resp})
}

func postShutdownServer(c *gin.Context) {
	restresponse.RestSuccess(c, gin.H{"status": "shutting down"})
	go func() {
		slog.Info("admin shutdown requested", slog.String("component", "server"))
		time.Sleep(200 * time.Millisecond)
		start := time.Now()
		n := eventspkg.ShutdownAllSSESessions()
		slog.Info("terminated SSE sessions for shutdown", slog.Int("count", n), slog.Duration("took", time.Since(start)))
		if err := restapi.ShutdownHttpServerWithTimeout(time.Second*5, true); err != nil {
			slog.Error("shutdown http", "error", err)
		}
		if err := restapi.ShutdownHttpsServerWithTimeout(time.Second*5, true); err != nil {
			slog.Error("shutdown https", "error", err)
		}
		os.Exit(0)
	}()
}

func postRestartNetworkServers(c *gin.Context) {
	restresponse.RestSuccess(c, gin.H{"status": "restarting-network"})
	go func() {
		slog.Info("admin network restart requested", slog.String("component", "server"))
		start := time.Now()
		n := eventspkg.ShutdownAllSSESessions()
		slog.Info("terminated SSE sessions for network restart", slog.Int("count", n), slog.Duration("took", time.Since(start)))
		restartNetworkListenersFromSettings()
	}()
}

func postRestartServer(c *gin.Context) {
	restresponse.RestSuccess(c, gin.H{"status": "restarting"})
	go func() {
		slog.Info("admin restart requested", slog.String("component", "server"))
		if err := updater.Restart(""); err != nil {
			slog.Error("failed to restart server", "error", err)
		}
	}()
}

func getBrandingSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	resp := gin.H{
		"brandLogoUrl": utilities.PtrVal(s.BrandLogoURL),
		"brandName":    utilities.PtrVal(s.BrandName),
	}
	restresponse.RestSuccess(c, resp)
}

func putBrandingSettings(c *gin.Context) {
	var body struct {
		BrandLogoURL string `json:"brandLogoUrl"`
		BrandName    string `json:"brandName"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	cxsettingspkg.CxSettings.BrandLogoURL = utilities.Ptr(strings.TrimSpace(body.BrandLogoURL))
	cxsettingspkg.CxSettings.BrandName = utilities.Ptr(strings.TrimSpace(body.BrandName))

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}
