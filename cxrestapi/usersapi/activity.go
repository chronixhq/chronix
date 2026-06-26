package usersapi

import (
	activitypkg "chronix/internal/activity"
	cxuserpkg "chronix/internal/cxuser"
	"chronix/internal/db/models"
	"fmt"
	"strconv"
	"time"

	"chronix/cxrestapi/apiutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func getMyActivity(c *gin.Context) {
	user := apiutil.UserFromGinContext(c)
	limit := 100
	if ls := c.Query("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil {
			limit = n
		}
	}

	items, err := activitypkg.GetUserActivity(user.ID, limit)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching activity")
		return
	}

	type item struct {
		ID      int64   `json:"id"`
		When    string  `json:"when"`
		Action  string  `json:"action"`
		Details *string `json:"details,omitempty"`
	}

	out := make([]item, 0, len(items))
	for _, it := range items {
		id := int64(0)
		if it.ID != nil {
			id = *it.ID
		}
		out = append(out, item{
			ID:      id,
			When:    it.CreatedAt.UTC().Format(time.RFC3339),
			Action:  it.Action,
			Details: it.Details,
		})
	}

	restresponse.RestSuccess(c, out)
}

func getAllUserActivity(c *gin.Context) {
	limit := 200
	if ls := c.Query("limit"); ls != "" {
		if n, err := strconv.Atoi(ls); err == nil {
			limit = n
		}
	}

	userID := int64(-1)
	if us := c.Query("user_id"); us != "" {
		if n, err := strconv.ParseInt(us, 10, 64); err == nil {
			userID = n
		}
	}

	var (
		items []*models.UserActivity
		err   error
	)
	if userID >= 0 {
		items, err = activitypkg.GetUserActivity(userID, limit)
	} else {
		items, err = activitypkg.GetAllUserActivity(limit)
	}
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching activity")
		return
	}

	userNames := map[int64]string{0: "Admin"}
	if users, err := cxuserpkg.UserList(); err == nil {
		for _, u := range users {
			userNames[u.ID] = u.Name
		}
	}

	type item struct {
		ID      int64   `json:"id"`
		UserID  int64   `json:"userId"`
		User    string  `json:"user"`
		When    string  `json:"when"`
		Action  string  `json:"action"`
		Details *string `json:"details,omitempty"`
	}

	out := make([]item, 0, len(items))
	for _, it := range items {
		id := int64(0)
		if it.ID != nil {
			id = *it.ID
		}
		name, ok := userNames[it.UserID]
		if !ok {
			name = fmt.Sprintf("User %d", it.UserID)
		}
		out = append(out, item{
			ID:      id,
			UserID:  it.UserID,
			User:    name,
			When:    it.CreatedAt.UTC().Format(time.RFC3339),
			Action:  it.Action,
			Details: it.Details,
		})
	}

	restresponse.RestSuccess(c, out)
}
