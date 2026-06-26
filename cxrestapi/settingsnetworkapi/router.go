package settingsnetworkapi

import "github.com/gin-gonic/gin"

type Config struct {
	DefaultHTTPPort  int
	DefaultHTTPSPort int
	DefaultAgentPort int
}

var defaults Config

func Register(app *gin.Engine, wrap func(gin.HandlerFunc) gin.HandlerFunc, cfg Config) {
	defaults = cfg

	app.GET("/settings/settings/http", wrap(getHTTPSettings))
	app.PUT("/settings/settings/http", wrap(putHTTPSettings))
	app.GET("/settings/settings/https", wrap(getHTTPSSettings))
	app.PUT("/settings/settings/https", wrap(putHTTPSSettings))
	app.GET("/settings/settings/agent", wrap(getAgentSettings))
	app.PUT("/settings/settings/agent", wrap(putAgentSettings))
	app.POST("/settings/settings/https/cert", wrap(uploadHTTPSCert))
	app.POST("/settings/settings/https/key", wrap(uploadHTTPSKey))
	app.POST("/settings/settings/https/upload", wrap(uploadHTTPSCertAndKey))
	app.DELETE("/settings/settings/https", wrap(deleteHTTPSCertAndKey))
	app.DELETE("/settings/settings/https/cert", wrap(deleteHTTPSCert))
	app.DELETE("/settings/settings/https/key", wrap(deleteHTTPSKey))
}
