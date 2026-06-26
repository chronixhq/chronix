package actionsapi

import "github.com/gin-gonic/gin"

func Register(app *gin.Engine) {
	app.GET("/actions", listActions)
	app.POST("/actions", createAction)
	app.GET("/actions/:id", getAction)
	app.PUT("/actions/:id", updateAction)
	app.PATCH("/actions/:id", patchAction)
	app.DELETE("/actions/:id", deleteAction)
	app.POST("/actions/validate-step", validateActionStep)
	app.POST("/actions/test", testAction)
}
