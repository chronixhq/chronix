package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/pkg/sqlutil"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type dbConnPayload struct {
	ID               *int64  `json:"id"`
	Name             *string `json:"name"`
	Driver           *string `json:"driver"`
	DSN              *string `json:"dsn"`
	Description      *string `json:"description"`
	AutoCheckEnabled *bool   `json:"autoCheckEnabled"`
	AutoCheckSeconds *int64  `json:"autoCheckSeconds"`
	AgentUUID        *string `json:"agentUuid"`
	AlertEmails      *string `json:"alertEmails"`
	AlertPhones      *string `json:"alertPhones"`
	NotifyOnFailure  *bool   `json:"notifyOnFailure"`
	Enabled          *bool   `json:"enabled"`
	Suspended        *bool   `json:"suspended"`
}

func validatePayload(p *dbConnPayload, isUpdate bool) error {
	if !isUpdate {
		if p.Name == nil || strings.TrimSpace(*p.Name) == "" {
			return errors.New("name is required")
		}
		if p.Driver == nil || *p.Driver == "" {
			return errors.New("driver is required")
		}
	} else {
		if p.Name != nil && strings.TrimSpace(*p.Name) == "" {
			return errors.New("name cannot be empty")
		}
		if p.Driver != nil && *p.Driver == "" {
			return errors.New("driver cannot be empty")
		}
	}

	if p.Driver != nil {
		drv := sqlutil.NormalizeDriver(strings.TrimSpace(*p.Driver))
		switch drv {
		case "postgres", "mysql", "mssql", "sqlserver", "sqlite", "oracle", "snowflake":
		default:
			return errors.New("unsupported driver: " + drv)
		}
	}
	return nil
}

func getAllConnections(c *gin.Context) {
	items, err := db.DbConnection.Order(db.DbConnection.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list connections", err.Error())
		return
	}
	uuids := make([]string, 0)
	uuidSeen := make(map[string]struct{})
	for _, it := range items {
		if it.AgentUUID != nil {
			u := strings.TrimSpace(*it.AgentUUID)
			if u != "" {
				if _, ok := uuidSeen[u]; !ok {
					uuidSeen[u] = struct{}{}
					uuids = append(uuids, u)
				}
			}
		}
	}
	agentNames := make(map[string]string, len(uuids))
	if len(uuids) > 0 {
		if agents, err := db.Agent.Where(db.Agent.UUID.In(uuids...)).Find(); err == nil {
			for _, a := range agents {
				agentNames[a.UUID] = a.Name
			}
		}
	}
	resp := make([]gin.H, 0, len(items))
	for _, it := range items {
		row, dsn := responseForConnection(it)
		if it.AgentUUID != nil {
			if name, ok := agentNames[strings.TrimSpace(*it.AgentUUID)]; ok && name != "" {
				row["agentName"] = name
			}
		}
		for k, v := range parseDSNBasic(it.Driver, dsn) {
			row[k] = v
		}
		resp = append(resp, row)
	}
	restresponse.RestSuccess(c, resp)
}

func getConnection(c *gin.Context) {
	item, err := db.DbConnection.Where(db.DbConnection.ID.Eq(atoi64(c.Param("id")))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	resp, dsn := responseForConnection(item)
	if item.AgentUUID != nil {
		if agent, err := db.Agent.Where(db.Agent.UUID.Eq(strings.TrimSpace(*item.AgentUUID))).First(); err == nil {
			resp["agentName"] = agent.Name
		}
	}
	for k, v := range parseDSNBasic(item.Driver, dsn) {
		resp[k] = v
	}
	restresponse.RestSuccess(c, resp)
}

func createDbConnection(c *gin.Context) {
	var p dbConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := validatePayload(&p, false); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}

	dsn := ""
	if p.DSN != nil {
		dsn = *p.DSN
	}
	encDSN, _ := secret.Encrypt(dsn)
	m := &models.DbConnection{
		Name:                     *p.Name,
		Driver:                   *p.Driver,
		Dsn:                      &encDSN,
		Description:              p.Description,
		AutoCheckEnabled:         utilities.Ptr(b2i(p.AutoCheckEnabled)),
		AutoCheckIntervalSeconds: utilities.Ptr(pickInt64Val(p.AutoCheckSeconds, 3600)),
		AgentUUID:                p.AgentUUID,
		AlertEmails:              p.AlertEmails,
		AlertPhones:              p.AlertPhones,
		NotifyOnFailure:          p.NotifyOnFailure,
		Enabled:                  utilities.Ptr(true),
		Suspended:                utilities.Ptr(false),
	}
	if err := db.DbConnection.Create(m); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to create connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created connection", fmt.Sprintf("%s (%s)", m.Name, m.Driver), c.ClientIP(), c.Request.UserAgent())
	resp, _ := responseForConnection(m)
	restresponse.RestSuccess(c, resp)
}

func updateConnection(c *gin.Context) {
	var p dbConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := validatePayload(&p, true); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}
	item, err := db.DbConnection.Where(db.DbConnection.ID.Eq(atoi64(c.Param("id")))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	if p.Name != nil {
		item.Name = *p.Name
	}
	if p.Driver != nil {
		item.Driver = *p.Driver
	}
	if p.DSN != nil {
		dsn := *p.DSN
		if item.Dsn != nil {
			oldDSN, _ := secret.Decrypt(*item.Dsn)
			driver := item.Driver
			if p.Driver != nil {
				driver = *p.Driver
			}
			dsn = mergeDSN(driver, dsn, oldDSN)
		}
		encDSN, _ := secret.Encrypt(dsn)
		item.Dsn = &encDSN
	}
	if p.Description != nil {
		item.Description = p.Description
	}
	if p.AutoCheckEnabled != nil {
		item.AutoCheckEnabled = utilities.Ptr(b2i(p.AutoCheckEnabled))
	}
	if p.AutoCheckSeconds != nil {
		item.AutoCheckIntervalSeconds = p.AutoCheckSeconds
	}
	if p.AgentUUID != nil {
		item.AgentUUID = p.AgentUUID
	}
	if p.AlertEmails != nil {
		item.AlertEmails = p.AlertEmails
	}
	if p.AlertPhones != nil {
		item.AlertPhones = p.AlertPhones
	}
	if p.NotifyOnFailure != nil {
		item.NotifyOnFailure = p.NotifyOnFailure
	}
	if p.Enabled != nil {
		if *p.Enabled && !utilities.PtrVal(item.Enabled) {
		}
		item.Enabled = p.Enabled
	}
	if p.Suspended != nil {
		item.Suspended = p.Suspended
	}
	if err := db.DbConnection.Save(item); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to update connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated connection", item.Name, c.ClientIP(), c.Request.UserAgent())
	resp, _ := responseForConnection(item)
	restresponse.RestSuccess(c, resp)
}

func deleteConnection(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.DbConnection.Where(db.DbConnection.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	usageCount, err := db.Job.Where(db.Job.ConnectionID.Eq(id), db.Job.Enabled.Is(true)).Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to check connection usage", err.Error())
		return
	}
	if usageCount > 0 {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot delete connection: it is referenced by enabled scheduled jobs")
		return
	}
	if _, err := db.DbConnection.Where(db.DbConnection.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted connection", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func duplicateConnection(c *gin.Context) {
	item, err := db.DbConnection.Where(db.DbConnection.ID.Eq(atoi64(c.Param("id")))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	newItem := *item
	newItem.ID = nil
	newItem.Name = "Copy Of " + item.Name
	newItem.CreatedAt = utilities.Ptr(time.Now())
	newItem.UpdatedAt = utilities.Ptr(time.Now())
	newItem.LastStatus = nil
	newItem.LastError = nil
	newItem.LastCheckedAt = nil
	if err := db.DbConnection.Create(&newItem); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to duplicate connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Duplicated connection", fmt.Sprintf("%s -> %s", item.Name, newItem.Name), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, newItem)
}

func responseForConnection(item *models.DbConnection) (gin.H, string) {
	dsn := ""
	if item.Dsn != nil {
		dsn, _ = secret.Decrypt(*item.Dsn)
	}
	return gin.H{
		"id":               item.ID,
		"name":             item.Name,
		"driver":           item.Driver,
		"description":      item.Description,
		"autoCheckEnabled": utilities.PtrVal(item.AutoCheckEnabled) != 0,
		"autoCheckSeconds": item.AutoCheckIntervalSeconds,
		"createdAt":        item.CreatedAt,
		"updatedAt":        item.UpdatedAt,
		"lastStatus":       item.LastStatus,
		"lastError":        item.LastError,
		"lastCheckedAt":    item.LastCheckedAt,
		"agentUuid":        item.AgentUUID,
		"alertEmails":      item.AlertEmails,
		"alertPhones":      item.AlertPhones,
		"notifyOnFailure":  item.NotifyOnFailure,
		"enabled":          item.Enabled,
		"suspended":        item.Suspended,
		"maskedDsn":        maskDSN(dsn),
	}, dsn
}
