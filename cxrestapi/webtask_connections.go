package cxrestapi

import (
	"chronix/cxrestapi/webtaskconnectionsapi"

	"github.com/gin-gonic/gin"
)

func webtaskConnectionsRouter(app *gin.Engine) {
	webtaskconnectionsapi.Register(app)
}
