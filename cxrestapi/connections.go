package cxrestapi

import "github.com/gin-gonic/gin"

func connectionsRouter(app *gin.Engine) {
	app.GET("/connections", getAllConnections)
	app.POST("/connections", createDbConnection)
	app.GET("/connections/:id", getConnection)
	app.PUT("/connections/:id", updateConnection)
	app.DELETE("/connections/:id", deleteConnection)
	app.POST("/connections/:id/test", testConnection)
	app.POST("/connections/:id/duplicate", duplicateConnection)
	app.POST("/connections/test", testConnectionFromDraft)
}
