package feedbackapi

import "github.com/gin-gonic/gin"

var wrapAdmin func(gin.HandlerFunc) gin.HandlerFunc

func Register(app *gin.Engine, adminWrapper func(gin.HandlerFunc) gin.HandlerFunc) {
	wrapAdmin = adminWrapper
	g := app.Group("/feedback")
	{
		g.POST("/bug-report", postBugReport)
		g.POST("/feature-request", postFeatureRequest)
		g.GET("/bug-reports", wrapAdmin(getBugReports))
		g.GET("/feature-requests", wrapAdmin(getFeatureRequests))
		g.GET("/attachments/:id", getFeedbackAttachment)

		g.PATCH("/bug-report/:id", wrapAdmin(patchBugReport))
		g.DELETE("/bug-report/:id", wrapAdmin(deleteBugReport))
		g.PATCH("/feature-request/:id", wrapAdmin(patchFeatureRequest))
		g.DELETE("/feature-request/:id", wrapAdmin(deleteFeatureRequest))
		g.POST("/bug-report/:id/attachments", wrapAdmin(postBugReportAttachments))
		g.POST("/feature-request/:id/attachments", wrapAdmin(postFeatureRequestAttachments))
		g.DELETE("/attachments/:id", wrapAdmin(deleteFeedbackAttachment))
	}
}
