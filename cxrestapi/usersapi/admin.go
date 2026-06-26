package usersapi

import (
	activitypkg "chronix/internal/activity"
	cxuserpkg "chronix/internal/cxuser"
	eventspkg "chronix/internal/events"
	"fmt"
	"strconv"

	"chronix/cxrestapi/apiutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func userList(c *gin.Context) {
	ul, err := cxuserpkg.UserList()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error getting user list", err.Error())
		return
	}
	restresponse.RestSuccess(c, ul)
}

func saveUser(c *gin.Context) {
	current := apiutil.UserFromGinContext(c)
	var newUser cxuserpkg.CxUser
	if err := c.BindJSON(&newUser); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Error parsing JSON", err.Error())
		return
	}
	if newUser.ID != 0 {
		newUser.Password = nil
	}

	var existing *cxuserpkg.CxUser
	if newUser.ID != 0 {
		if users, err := cxuserpkg.UserList(); err == nil {
			for _, u := range users {
				if u.ID == newUser.ID {
					existing = u
					break
				}
			}
		}
	}

	adminDesired := newUser.Admin
	enabledDesired := newUser.Enabled

	if existing != nil && existing.Admin && existing.Enabled && (!adminDesired || !enabledDesired) {
		if users, err := cxuserpkg.UserList(); err == nil {
			otherEnabledAdmins := 0
			for _, u := range users {
				if u.ID != newUser.ID && u.Admin && u.Enabled {
					otherEnabledAdmins++
				}
			}
			if otherEnabledAdmins == 0 {
				msg := "Cannot disable or revoke admin from the last enabled admin user"
				if newUser.ID == current.ID {
					msg = "Cannot remove admin or disable your own account because no other enabled admin users exist"
				}
				restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, msg)
				return
			}
		}
	}

	if existing != nil && existing.Admin && !adminDesired {
		if users, err := cxuserpkg.UserList(); err == nil {
			otherAdmins := 0
			for _, u := range users {
				if u.ID != newUser.ID && u.Admin {
					otherAdmins++
				}
			}
			if otherAdmins == 0 {
				restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot revoke admin role from the last admin user")
				return
			}
		}
	}

	if err := newUser.Save(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving user", err.Error())
		return
	}

	if adminDesired != newUser.Admin {
		newUser.SetAdmin(adminDesired)
		if !adminDesired {
			hooks.RevokeUserAuthKeys(newUser.ID)
		}
		action := "Granted admin role"
		if !adminDesired {
			action = "Revoked admin role"
		}
		details := fmt.Sprintf("User: %s (%s)", newUser.Name, newUser.Email)
		_ = activitypkg.RecordUserActivity(current.ID, action, details, c.ClientIP(), c.Request.UserAgent())
	}

	if enabledDesired != newUser.Enabled {
		newUser.SetEnabled(enabledDesired)
		if !enabledDesired {
			hooks.RevokeUserAuthKeys(newUser.ID)
		}
		action := "Enabled user"
		if !enabledDesired {
			action = "Disabled user"
		}
		details := fmt.Sprintf("User: %s (%s)", newUser.Name, newUser.Email)
		_ = activitypkg.RecordUserActivity(current.ID, action, details, c.ClientIP(), c.Request.UserAgent())
	}

	requiresLogout := false
	reason := ""
	if existing != nil {
		if existing.Admin && !adminDesired {
			requiresLogout = true
			reason = "Your admin access was revoked by an administrator."
		}
		if existing.Enabled && !enabledDesired {
			requiresLogout = true
			reason = "Your account was disabled by an administrator."
		}
		if existing.Email != "" && existing.Email != newUser.Email {
			requiresLogout = true
			reason = "Your email address was changed by an administrator. Please sign in again."
			hooks.RevokeUserAuthKeys(newUser.ID)
		}
	}

	restresponse.RestSuccess(c, newUser)

	_ = eventspkg.SendUserIDEvent(newUser.ID, eventspkg.SSEEventUserUpdate, eventspkg.UserUpdateEvent{
		ID:      newUser.ID,
		Email:   newUser.Email,
		Name:    newUser.Name,
		Phone:   newUser.Phone,
		Admin:   adminDesired,
		Enabled: enabledDesired,
	})
	if requiresLogout {
		_ = eventspkg.SendUserIDEvent(newUser.ID, eventspkg.SSEEventLogout, eventspkg.LogoutEvent{Reason: reason})
	}
}

func checkEmailForMe(c *gin.Context) {
	current := apiutil.UserFromGinContext(c)
	email := c.Query("email")
	exclude := current.ID
	if ex := c.Query("excludeId"); ex != "" {
		if v, err := strconv.ParseInt(ex, 10, 64); err == nil {
			exclude = v
		}
	}
	ok, _ := cxuserpkg.IsEmailAvailable(email, exclude)
	restresponse.RestSuccess(c, gin.H{"available": ok})
}

func checkEmailGeneric(c *gin.Context) {
	email := c.Query("email")
	exclude := int64(0)
	if ex := c.Query("excludeId"); ex != "" {
		if v, err := strconv.ParseInt(ex, 10, 64); err == nil {
			exclude = v
		}
	}
	ok, _ := cxuserpkg.IsEmailAvailable(email, exclude)
	restresponse.RestSuccess(c, gin.H{"available": ok})
}

func changeUserPassword(c *gin.Context) {
	admin := apiutil.UserFromGinContext(c)
	idParam := c.Param("id")
	if idParam == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Missing user id")
		return
	}
	uid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || uid <= 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid user id")
		return
	}

	var body struct {
		NewPassword string `json:"newPassword"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if len(body.NewPassword) < 8 {
		restresponse.RestErrorRespond(c, restresponse.InvalidArgument, "Password must be at least 8 characters")
		return
	}

	target, err := cxuserpkg.GetUserByID(uid)
	if err != nil || target == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "User not found")
		return
	}

	upd := cxuserpkg.CxUser{CxUser: target.CxUser}
	upd.Password = &body.NewPassword
	if err := upd.Save(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error updating password", err.Error())
		return
	}

	hooks.RevokeUserAuthKeys(uid)
	details := fmt.Sprintf("User: %s (%s)", target.Name, target.Email)
	_ = activitypkg.RecordUserActivity(admin.ID, "Changed user password (Admin)", details, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteUser(c *gin.Context) {
	admin := apiutil.UserFromGinContext(c)
	idParam := c.Param("id")
	if idParam == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Missing user id")
		return
	}
	uid, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil || uid <= 0 {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid user id")
		return
	}
	if uid == admin.ID {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "You cannot delete your own account")
		return
	}

	target, err := cxuserpkg.GetUserByID(uid)
	if err != nil || target == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "User not found")
		return
	}

	if target.Admin {
		if users, err := cxuserpkg.UserList(); err == nil {
			otherAdmins := 0
			for _, u := range users {
				if u.ID != uid && u.Admin {
					otherAdmins++
				}
			}
			if otherAdmins == 0 {
				restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot delete the last admin user")
				return
			}
		}
	}

	if err := target.Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error deleting user", err.Error())
		return
	}
	hooks.RevokeUserAuthKeys(uid)
	details := fmt.Sprintf("User: %s (%s)", target.Name, target.Email)
	_ = activitypkg.RecordUserActivity(admin.ID, "Deleted user", details, c.ClientIP(), c.Request.UserAgent())
	_ = eventspkg.SendUserIDEvent(uid, eventspkg.SSEEventLogout, eventspkg.LogoutEvent{Reason: "Your account was deleted by an administrator."})
	restresponse.RestSuccessNoContent(c)
}
