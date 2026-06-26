package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	notifypkg "chronix/internal/notify"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func notificationsRouter(r *gin.Engine) {
	r.GET("/notifications/recent", getRecentNotifications)
	r.GET("/notifications", listNotifications)
	r.POST("/notifications/mark-seen", markSeenNotifications)
	r.POST("/notifications/mark-removed", markRemovedNotifications)
}

func getRecentNotifications(c *gin.Context) {
	user := userFromGinContext(c)
	limit := 20
	if ls := c.Query("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil {
			limit = n
		}
	}
	items, err := notifypkg.GetRecentForUser(user.ID, limit)
	if err != nil {
		slog.Error("get recent notifications", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching notifications")
		return
	}
	count, err := notifypkg.CountUnseenForUser(user.ID)
	if err != nil {
		count = 0
	}
	restresponse.RestSuccess(c, gin.H{"items": items, "unseenCount": count})
}

func listNotifications(c *gin.Context) {
	user := userFromGinContext(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "50"))
	category := c.Query("category")
	severity := c.Query("severity")
	var seenPtr *bool
	if sv := c.Query("seen"); sv != "" {
		b := sv == "true" || sv == "1"
		seenPtr = &b
	}
	items, total, err := notifypkg.ListForUser(user.ID, page, pageSize, notifypkg.NotificationCategory(category), notifypkg.NotificationSeverity(severity), seenPtr)
	if err != nil {
		slog.Error("list notifications", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching notifications")
		return
	}
	restresponse.RestSuccess(c, gin.H{"items": items, "total": total, "page": page, "pageSize": pageSize})
}

func markSeenNotifications(c *gin.Context) {
	user := userFromGinContext(c)
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := notifypkg.MarkSeen(user.ID, body.IDs); err != nil {
		slog.Error("mark seen", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error marking seen")
		return
	}
	_ = activitypkg.RecordUserActivity(user.ID, "Marked notifications as read", fmt.Sprintf("Marked %d notifications as read", len(body.IDs)), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func markRemovedNotifications(c *gin.Context) {
	user := userFromGinContext(c)
	var body struct {
		IDs []int64 `json:"ids"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := notifypkg.MarkRemoved(user.ID, body.IDs); err != nil {
		slog.Error("mark removed", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error marking removed")
		return
	}
	_ = activitypkg.RecordUserActivity(user.ID, "Removed notifications", fmt.Sprintf("Removed %d notifications", len(body.IDs)), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}
