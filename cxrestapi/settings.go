package cxrestapi

import (
	"chronix/cxrestapi/settingsnetworkapi"

	"github.com/gin-gonic/gin"
)

func settingsRouter(utApp *gin.Engine) {
	utApp.POST("/serverUrl", adminFunc(setServerURL))

	utApp.GET("/settings/settings/server-url", adminFunc(getServerURL))
	utApp.PUT("/settings/settings/server-url", adminFunc(putServerURL))

	utApp.GET("/settings/settings/email", adminFunc(getEmailSettings))
	utApp.PUT("/settings/settings/email", adminFunc(putEmailSettings))
	utApp.POST("/settings/settings/email/test", adminFunc(postTestEmailSettings))

	settingsnetworkapi.Register(utApp, adminFunc, settingsnetworkapi.Config{
		DefaultHTTPPort:  DefaultHTTPPort,
		DefaultHTTPSPort: DefaultHTTPSPort,
		DefaultAgentPort: DefaultAgentPort,
	})

	utApp.GET("/settings/settings/sms", adminFunc(getSMSSettings))
	utApp.PUT("/settings/settings/sms", adminFunc(putSMSSettings))
	utApp.POST("/settings/settings/sms/test", adminFunc(postTestSMSSettings))

	utApp.GET("/settings/settings/alerts", adminFunc(getGlobalAlertSettings))
	utApp.PUT("/settings/settings/alerts", adminFunc(putGlobalAlertSettings))

	utApp.GET("/settings/settings/branding", adminFunc(getBrandingSettings))
	utApp.PUT("/settings/settings/branding", adminFunc(putBrandingSettings))

	utApp.GET("/settings/settings/updater", adminFunc(getUpdaterSettings))
	utApp.PUT("/settings/settings/updater/app", adminFunc(putUpdaterAppSettings))
	utApp.PUT("/settings/settings/updater/agent", adminFunc(putUpdaterAgentSettings))

	utApp.GET("/settings/settings/summary", adminFunc(getSettingsSummary))
	utApp.GET("/settings/settings/active", adminFunc(getActiveSettings))
	utApp.POST("/settings/restart-network", adminFunc(postRestartNetworkServers))
	utApp.POST("/settings/restart", adminFunc(postRestartServer))
	utApp.POST("/settings/shutdown", adminFunc(postShutdownServer))
}
