package usersapi

import (
	cxuserpkg "chronix/internal/cxuser"

	"github.com/gin-gonic/gin"
)

type Hooks struct {
	WrapAdmin          func(gin.HandlerFunc) gin.HandlerFunc
	RefreshAuthCookie  func(*gin.Context, *cxuserpkg.CxUser)
	FullLogout         func(*gin.Context)
	RevokeUserAuthKeys func(int64)
}

var hooks Hooks

func Register(app *gin.Engine, h Hooks) {
	hooks = h

	app.POST("/user", hooks.WrapAdmin(saveUser))
	app.GET("/users", hooks.WrapAdmin(userList))
	app.GET("/settings/users/activity", hooks.WrapAdmin(getAllUserActivity))
	app.GET("/users/check-email", hooks.WrapAdmin(checkEmailGeneric))
	app.PUT("/settings/users/:id/password", hooks.WrapAdmin(changeUserPassword))
	app.DELETE("/settings/users/:id", hooks.WrapAdmin(deleteUser))

	app.GET("/me", getMe)
	app.PUT("/me", updateMe)
	app.PUT("/me/password", changeMyPassword)
	app.GET("/me/activity", getMyActivity)
	app.GET("/me/check-email", checkEmailForMe)
}
