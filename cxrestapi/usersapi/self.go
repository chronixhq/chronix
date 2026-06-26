package usersapi

import (
	activitypkg "chronix/internal/activity"
	cxuserpkg "chronix/internal/cxuser"

	"chronix/cxrestapi/apiutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func getMe(c *gin.Context) {
	uc := apiutil.UserFromGinContext(c)
	if uc.ID == 0 {
		uc.Name = "Setup Admin"
		uc.Password = nil
		restresponse.RestSuccess(c, uc)
		return
	}

	full, err := cxuserpkg.GetUserByID(uc.ID)
	if err != nil || full == nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error loading user")
		return
	}
	full.Password = nil
	full.AuthKey = uc.AuthKey
	restresponse.RestSuccess(c, full)
}

func updateMe(c *gin.Context) {
	current := apiutil.UserFromGinContext(c)
	if current.ID == 0 {
		restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Setup admin cannot update profile")
		return
	}

	var body cxuserpkg.CxUser
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Error parsing JSON", err.Error())
		return
	}

	update := cxuserpkg.CxUser{CxUser: body.CxUser}
	update.ID = current.ID
	update.Admin = current.Admin
	update.Enabled = current.Enabled
	update.Password = nil
	if err := update.Save(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving profile", err.Error())
		return
	}

	update.AuthKey = current.AuthKey
	hooks.RefreshAuthCookie(c, &update)
	if c.IsAborted() {
		return
	}

	details := "name=" + update.Name + ", email=" + update.Email
	_ = activitypkg.RecordUserActivity(current.ID, "Updated profile", details, c.ClientIP(), c.Request.UserAgent())
	update.Password = nil
	restresponse.RestSuccess(c, update)
}

func changeMyPassword(c *gin.Context) {
	current := apiutil.UserFromGinContext(c)
	if current.ID == 0 {
		restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Setup admin cannot change password")
		return
	}

	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if len(body.CurrentPassword) == 0 || len(body.NewPassword) == 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Current and new password required")
		return
	}

	if _, err := cxuserpkg.Login(current.Email, body.CurrentPassword); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid current password")
		return
	}

	upd := cxuserpkg.CxUser{CxUser: current.CxUser}
	upd.Password = &body.NewPassword
	upd.ForcePasswordChange = false
	if err := upd.Save(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error updating password", err.Error())
		return
	}

	_ = activitypkg.RecordUserActivity(current.ID, "Changed password", "", c.ClientIP(), c.Request.UserAgent())
	hooks.FullLogout(c)
}
