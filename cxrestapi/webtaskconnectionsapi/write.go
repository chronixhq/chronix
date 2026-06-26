package webtaskconnectionsapi

import (
	"chronix/cxrestapi/apiutil"
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"errors"
	"fmt"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func createWebtaskConnection(c *gin.Context) {
	var p webtaskConnPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}
	if err := validateWebtaskConnPayload(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}

	at := ""
	if p.AuthType != nil {
		at = *p.AuthType
	}
	if p.AuthConfig != nil {
		switch at {
		case "basic":
			if pass, ok := p.AuthConfig["password"].(string); ok && pass != "" {
				p.AuthConfig["password"], _ = secret.Encrypt(pass)
			}
		case "bearer":
			if token, ok := p.AuthConfig["token"].(string); ok && token != "" {
				p.AuthConfig["token"], _ = secret.Encrypt(token)
			}
		case "header":
			if val, ok := p.AuthConfig["header_value"].(string); ok && val != "" {
				p.AuthConfig["header_value"], _ = secret.Encrypt(val)
			}
		}
	}

	authConfig := datatypes.JSONMap(p.AuthConfig)
	row := &models.WebtaskConnection{
		Name:                     *p.Name,
		Description:              p.Description,
		BaseURL:                  p.BaseURL,
		AuthType:                 at,
		AuthConfig:               &authConfig,
		AgentUUID:                p.AgentUUID,
		AutoCheckEnabled:         utilities.Ptr(AnyToI64(p.AutoCheckEnabled, 1)),
		AutoCheckIntervalSeconds: utilities.Ptr(AnyToI64(p.AutoCheckSeconds, 300)),
		AlertEmails:              p.AlertEmails,
		AlertPhones:              p.AlertPhones,
		NotifyOnFailure:          p.NotifyOnFailure,
		Enabled:                  utilities.Ptr(true),
		Suspended:                utilities.Ptr(false),
		CreatedAt:                utilities.Ptr(time.Now()),
		UpdatedAt:                utilities.Ptr(time.Now()),
	}

	if err := db.WebtaskConnection.Create(row); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to create connection", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created webtask connection", row.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, MapWebtaskConnection(row))
}

func updateWebtaskConnection(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	old, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "connection not found")
		return
	}

	var p webtaskConnPayload
	if err := c.ShouldBindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid payload", err.Error())
		return
	}

	at := old.AuthType
	if p.AuthType != nil {
		at = *p.AuthType
	}
	if p.AuthConfig != nil {
		switch at {
		case "basic":
			if pass, ok := p.AuthConfig["password"].(string); ok && pass != "" {
				if pass == "<redacted>" {
					if old.AuthConfig != nil {
						p.AuthConfig["password"] = (*old.AuthConfig)["password"]
					}
				} else {
					p.AuthConfig["password"], _ = secret.Encrypt(pass)
				}
			}
		case "bearer":
			if token, ok := p.AuthConfig["token"].(string); ok && token != "" {
				if token == "<redacted>" {
					if old.AuthConfig != nil {
						p.AuthConfig["token"] = (*old.AuthConfig)["token"]
					}
				} else {
					p.AuthConfig["token"], _ = secret.Encrypt(token)
				}
			}
		case "header":
			if val, ok := p.AuthConfig["header_value"].(string); ok && val != "" {
				if val == "<redacted>" {
					if old.AuthConfig != nil {
						p.AuthConfig["header_value"] = (*old.AuthConfig)["header_value"]
					}
				} else {
					p.AuthConfig["header_value"], _ = secret.Encrypt(val)
				}
			}
		}
		authConfig := datatypes.JSONMap(p.AuthConfig)
		old.AuthConfig = &authConfig
	}

	if p.Name != nil {
		old.Name = *p.Name
	}
	if p.Description != nil {
		old.Description = p.Description
	}
	if p.BaseURL != nil {
		old.BaseURL = p.BaseURL
	}
	if p.AuthType != nil {
		old.AuthType = *p.AuthType
	}
	if p.AgentUUID != nil {
		old.AgentUUID = p.AgentUUID
	}
	if p.AutoCheckEnabled != nil {
		old.AutoCheckEnabled = utilities.Ptr(AnyToI64(p.AutoCheckEnabled, 1))
	}
	if p.AutoCheckSeconds != nil {
		old.AutoCheckIntervalSeconds = utilities.Ptr(AnyToI64(p.AutoCheckSeconds, 300))
	}
	if p.AlertEmails != nil {
		old.AlertEmails = p.AlertEmails
	}
	if p.AlertPhones != nil {
		old.AlertPhones = p.AlertPhones
	}
	if p.NotifyOnFailure != nil {
		old.NotifyOnFailure = p.NotifyOnFailure
	}
	if p.Enabled != nil {
		if *p.Enabled && (old.Enabled == nil || !*old.Enabled) {
		}
		old.Enabled = p.Enabled
	}
	if p.Suspended != nil {
		old.Suspended = p.Suspended
	}
	old.UpdatedAt = utilities.Ptr(time.Now())

	if _, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).Updates(old); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to update connection", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated webtask connection", old.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, MapWebtaskConnection(old))
}

func deleteWebtaskConnection(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	item, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "connection not found")
		return
	}
	if _, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to delete connection", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted webtask connection", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func duplicateWebtaskConnection(c *gin.Context) {
	id := apiutil.Atoi64(c.Param("id"))
	item, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Webtask connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to load webtask connection", err.Error())
		return
	}

	now := time.Now()
	newItem := *item
	newItem.ID = nil
	newItem.Name = "Copy Of " + item.Name
	newItem.CreatedAt = &now
	newItem.UpdatedAt = &now
	newItem.LastStatus = nil
	newItem.LastError = nil
	newItem.LastCheckedAt = nil

	if err := db.WebtaskConnection.Create(&newItem); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to duplicate webtask connection", err.Error())
		return
	}

	user := apiutil.UserFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Duplicated webtask connection", fmt.Sprintf("%s -> %s", item.Name, newItem.Name), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, MapWebtaskConnection(&newItem))
}
