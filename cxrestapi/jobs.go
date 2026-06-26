package cxrestapi

import "github.com/gin-gonic/gin"

func jobsRouter(app *gin.Engine) {
	app.GET("/jobs", listJobs)
	app.POST("/jobs", createJob)
	app.GET("/jobs/:id", getJob)
	app.PUT("/jobs/:id", updateJob)
	app.DELETE("/jobs/:id", deleteJob)
	app.POST("/jobs/:id/enable", enableJob)
	app.POST("/jobs/:id/disable", disableJob)
	app.POST("/jobs/:id/runNow", runNowJob)
	app.GET("/jobs/:id/runs", listJobRuns)
}
