package cxrestapi

import "github.com/gin-gonic/gin"

func shellConnectionsRouter(app *gin.Engine) {
	g := app.Group("/shell")
	g.GET("/connections", listShellConnections)
	g.POST("/connections", createShellConnection)
	g.GET("/connections/:id", getShellConnection)
	g.PUT("/connections/:id", updateShellConnection)
	g.DELETE("/connections/:id", deleteShellConnection)
	g.POST("/connections/:id/test", testShellConnection)
	g.POST("/connections/:id/duplicate", duplicateShellConnection)
	g.POST("/connections/:id/clear-secret", clearShellConnectionSecret)
	g.POST("/connections/test", testShellConnectionFromDraft)
	g.POST("/connections/generate-keypair", generateKeyPair)
}
