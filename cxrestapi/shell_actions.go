package cxrestapi

import (
	"chronix/cxrestapi/shellactionsapi"

	"github.com/gin-gonic/gin"
)

func shellActionsRouter(app *gin.Engine) {
	shellactionsapi.Register(app)
}
