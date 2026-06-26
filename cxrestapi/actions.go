package cxrestapi

import (
	"chronix/cxrestapi/actionsapi"

	"github.com/gin-gonic/gin"
)

func actionsRouter(app *gin.Engine) {
	actionsapi.Register(app)
}
