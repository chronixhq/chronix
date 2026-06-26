package cxrestapi

import (
	"chronix/cxrestapi/feedbackapi"
	serverruntime "chronix/internal/serverruntime"

	"github.com/gin-gonic/gin"
)

func feedbackRouter(app *gin.Engine) {
	if !serverruntime.FeedbackEnabled {
		return
	}
	feedbackapi.Register(app, adminFunc)
}
