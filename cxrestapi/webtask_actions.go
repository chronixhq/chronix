package cxrestapi

import (
	"chronix/cxrestapi/webtaskactionsapi"

	"github.com/gin-gonic/gin"
)

func webtaskActionsRouter(app *gin.Engine) {
	webtaskactionsapi.Register(app)
}
