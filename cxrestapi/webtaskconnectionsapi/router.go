package webtaskconnectionsapi

import "github.com/gin-gonic/gin"

func Register(app *gin.Engine) {
	rg := app.Group("/connections/webtask")
	rg.GET("", listWebtaskConnections)
	rg.POST("", createWebtaskConnection)
	rg.POST("/test", testWebtaskConnectionFromDraft)
	rg.GET("/:id", getWebtaskConnection)
	rg.PUT("/:id", updateWebtaskConnection)
	rg.DELETE("/:id", deleteWebtaskConnection)
	rg.POST("/:id/test", testWebtaskConnection)
	rg.POST("/:id/duplicate", duplicateWebtaskConnection)
}
