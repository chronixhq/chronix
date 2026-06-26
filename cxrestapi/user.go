package cxrestapi

import (
	"chronix/cxrestapi/usersapi"
	cxuserpkg "chronix/internal/cxuser"

	"github.com/gin-gonic/gin"
)

func userRouter(utApp *gin.Engine) {
	usersapi.Register(utApp, usersapi.Hooks{
		WrapAdmin: adminFunc,
		RefreshAuthCookie: func(c *gin.Context, user *cxuserpkg.CxUser) {
			refreshAuthCookie(c, user)
		},
		FullLogout: fullLogout,
		RevokeUserAuthKeys: func(userID int64) {
			RevokeUserAuthKeys(userID)
		},
	})
}
