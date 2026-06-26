package cxrestapi

import (
	"chronix/internal/db"
	"chronix/internal/secret"
	"chronix/internal/sqlrunner"
	"chronix/pkg/sqlutil"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func testConnection(c *gin.Context) {
	item, err := db.DbConnection.Where(db.DbConnection.ID.Eq(atoi64(c.Param("id")))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	dsn := ""
	if item.Dsn != nil {
		dsn, _ = secret.Decrypt(*item.Dsn)
	}
	ok, msg := testConnectionViaRunner(c, item.Driver, dsn, item.AgentUUID)
	now := time.Now().UTC()
	status := "ok"
	var lastErr *string
	if !ok {
		status = "error"
		lastErr = &msg
	}
	item.LastStatus = &status
	item.LastError = lastErr
	item.LastCheckedAt = &now
	_ = db.DbConnection.Save(item)
	restresponse.RestSuccess(c, gin.H{"ok": ok, "message": msg})
}

func testConnectionFromDraft(c *gin.Context) {
	var p dbConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if p.ID != nil && *p.ID > 0 {
		rec, err := db.DbConnection.Where(db.DbConnection.ID.Eq(*p.ID)).First()
		if err == nil && rec != nil && rec.Dsn != nil {
			decDSN, _ := secret.Decrypt(*rec.Dsn)
			driver := ""
			if p.Driver != nil {
				driver = *p.Driver
			}
			dsn := ""
			if p.DSN != nil {
				dsn = *p.DSN
			}
			merged := mergeDSN(driver, dsn, decDSN)
			p.DSN = &merged
		}
	}
	driver := ""
	if p.Driver != nil {
		driver = *p.Driver
	}
	dsn := ""
	if p.DSN != nil {
		dsn = *p.DSN
	}
	ok, msg := testConnectionViaRunner(c, driver, dsn, p.AgentUUID)
	restresponse.RestSuccess(c, gin.H{"ok": ok, "message": msg})
}

func mergeDSN(driver, draft, old string) string {
	drv := sqlutil.NormalizeDriver(driver)
	if drv == "sqlite" {
		return draft
	}
	uDraft, err := url.Parse(draft)
	if err != nil {
		return draft
	}
	uOld, err := url.Parse(old)
	if err != nil {
		return draft
	}
	if drv == "postgres" || drv == "mysql" || drv == "oracle" || drv == "sqlserver" || drv == "snowflake" {
		if uDraft.User != nil {
			if _, has := uDraft.User.Password(); has {
				return draft
			}
			if uOld.User != nil {
				if pass, has := uOld.User.Password(); has {
					uDraft.User = url.UserPassword(uDraft.User.Username(), pass)
					return uDraft.String()
				}
			}
		}
	}
	if drv == "sqlserver" {
		qDraft := uDraft.Query()
		if qDraft.Get("Password") != "" {
			return draft
		}
		qOld := uOld.Query()
		if p := qOld.Get("Password"); p != "" {
			qDraft.Set("Password", p)
			uDraft.RawQuery = qDraft.Encode()
			return uDraft.String()
		}
	}
	return draft
}

func testConnectionViaRunner(c *gin.Context, driver, dsn string, agentUUID *string) (bool, string) {
	var runner sqlrunner.SQLRunner
	if agentUUID != nil && strings.TrimSpace(*agentUUID) != "" {
		runner = sqlrunner.AgentRunner{AgentID: strings.TrimSpace(*agentUUID)}
	} else {
		runner = sqlrunner.LocalRunner{}
	}
	drv := sqlutil.NormalizeDriver(driver)
	var probe string
	switch drv {
	case "postgres", "pgx", "mysql", "mssql", "sqlserver", "sqlite", "oracle", "snowflake":
		probe = "SELECT 1"
	default:
		return false, fmt.Sprintf("driver %s not supported yet", drv)
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	rows, _, err := runner.Query(ctx, drv, dsn, probe, nil, 1)
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "agent not connected") {
			return false, "agent not connected"
		}
		return false, msg
	}
	if rows >= 0 {
		return true, "connection ok"
	}
	return true, "connection ok"
}
