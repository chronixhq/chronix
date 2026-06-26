package webtaskactionsapi

import "github.com/gin-gonic/gin"

func Register(app *gin.Engine) {
	rg := app.Group("/actions/webtask")
	rg.GET("", listWebtaskActions)
	rg.POST("", createWebtaskAction)
	rg.GET("/:id", getWebtaskAction)
	rg.PUT("/:id", updateWebtaskAction)
	rg.PATCH("/:id", patchWebtaskAction)
	rg.DELETE("/:id", deleteWebtaskAction)
	rg.POST("/test", testWebtaskAction)
}
