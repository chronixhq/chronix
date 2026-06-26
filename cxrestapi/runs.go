package cxrestapi

import "github.com/gin-gonic/gin"

func runsRouter(app *gin.Engine) {
	app.GET("/runs", listRuns)
	app.GET("/runs/:runId", getRunDetail)
	app.GET("/runs/:runId/progress", getRunProgress)
	app.POST("/runs/:runId/cancel", cancelRun)
	app.POST("/runs/:runId/rerun", rerunRun)
}
