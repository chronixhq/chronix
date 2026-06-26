package shellactionsapi

import "github.com/gin-gonic/gin"

func Register(app *gin.Engine) {
	g := app.Group("/shell")
	g.GET("/actions", listShellActions)
	g.POST("/actions", createShellAction)
	g.GET("/actions/:id", getShellAction)
	g.PUT("/actions/:id", updateShellAction)
	g.PATCH("/actions/:id", patchShellAction)
	g.DELETE("/actions/:id", deleteShellAction)
	g.POST("/actions/validate", validateShellScript)
	g.POST("/actions/test", testShellAction)
}
