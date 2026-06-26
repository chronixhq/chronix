package cxrestapi

import (
	"chronix/pkg/sqlutil"
	"fmt"
	"net/url"
	"strings"
)

func maskDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err == nil && u != nil {
		if u.User != nil {
			if _, hasPw := u.User.Password(); hasPw {
				u.User = url.UserPassword(u.User.Username(), "***")
			}
		}
		q := u.Query()
		if q.Has("password") {
			q.Set("password", "***")
			u.RawQuery = q.Encode()
		}
		return u.String()
	}
	dsn = strings.ReplaceAll(dsn, "://", "://***:***@")
	dsn = strings.ReplaceAll(dsn, "password=", "password=***")
	return dsn
}

func parseDSNBasic(driver, dsn string) map[string]any {
	out := map[string]any{}
	drv := sqlutil.NormalizeDriver(driver)
	switch drv {
	case "sqlite":
		path := dsn
		if strings.HasPrefix(dsn, "file:") {
			path = strings.TrimPrefix(dsn, "file:")
		}
		out["filePath"] = path
		out["host"] = path
	case "postgres", "mysql", "mssql", "sqlserver", "oracle", "snowflake":
		if u, err := url.Parse(dsn); err == nil {
			if h := u.Hostname(); h != "" {
				out["host"] = h
			}
			if p := u.Port(); p != "" {
				out["port"] = p
			}
			path := strings.TrimPrefix(u.Path, "/")
			if path != "" {
				out["database"] = path
			} else {
				q := u.Query()
				if db := q.Get("database"); db != "" {
					out["database"] = db
				} else if db := q.Get("Database"); db != "" {
					out["database"] = db
				}
			}
			if u.User != nil {
				out["username"] = u.User.Username()
				if _, has := u.User.Password(); has {
					out["hasPassword"] = true
				}
			}
			q := u.Query()
			if q.Get("password") != "" || q.Get("Password") != "" {
				out["hasPassword"] = true
			}
		}
	}
	return out
}

func atoi64(s string) int64 {
	var v int64
	_, _ = fmt.Sscan(s, &v)
	return v
}

func pickBool(a *bool, def bool) bool {
	if a != nil {
		return *a
	}
	return def
}

func pickInt64Val(a *int64, def int64) int64 {
	if a != nil {
		return *a
	}
	return def
}

func b2i(a *bool) int64 {
	if a != nil && *a {
		return 1
	}
	return 0
}
