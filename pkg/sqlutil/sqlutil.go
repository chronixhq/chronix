// Package sqlutil provides common utilities for working with database SQL.
package sqlutil

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// NormalizeDriver returns a standardized driver name for a given alias.
func NormalizeDriver(d string) string {
	s := strings.ToLower(d)
	if s == "pgx" || s == "postgresql" || s == "cockroachdb" || s == "cockroach" {
		return "postgres"
	}
	if s == "mssql" {
		return "sqlserver"
	}
	if s == "tidb" || s == "mariadb" {
		return "mysql"
	}
	return s
}

// NormalizeDSN converts a connection string into a format recognized by the underlying driver.
func NormalizeDSN(driver, dsn string) string {
	drv := NormalizeDriver(driver)
	if drv == "mysql" && (strings.HasPrefix(dsn, "mysql://") || strings.HasPrefix(dsn, "mariadb://")) {
		u, err := url.Parse(dsn)
		if err != nil {
			return dsn
		}
		userinfo := ""
		if u.User != nil {
			userinfo = u.User.Username()
			if p, ok := u.User.Password(); ok {
				userinfo += ":" + p
			}
			userinfo += "@"
		}
		host := u.Host
		if host == "" {
			host = "localhost:3306"
		}
		db := strings.TrimPrefix(u.Path, "/")
		query := u.RawQuery

		newDsn := userinfo + "tcp(" + host + ")/" + db
		if query != "" {
			newDsn += "?" + query
		}
		return newDsn
	}
	if drv == "snowflake" && strings.HasPrefix(dsn, "snowflake://") {
		u, err := url.Parse(dsn)
		if err != nil {
			return dsn
		}
		userinfo := ""
		if u.User != nil {
			userinfo = u.User.Username()
			if p, ok := u.User.Password(); ok {
				userinfo += ":" + p
			}
			userinfo += "@"
		}
		account := u.Host
		path := u.Path
		query := u.RawQuery

		newDsn := userinfo + account + path
		if query != "" {
			newDsn += "?" + query
		}
		return newDsn
	}
	return dsn
}

// Placeholder returns the appropriate SQL parameter placeholder for the given driver.
func Placeholder(driver string, index int) string {
	drv := NormalizeDriver(driver)
	switch drv {
	case "postgres":
		return fmt.Sprintf("$%d", index)
	case "mysql", "sqlite":
		return "?"
	case "sqlserver":
		return fmt.Sprintf("@p%d", index)
	case "oracle":
		return fmt.Sprintf(":%d", index)
	default:
		return "?"
	}
}

// IsInsert returns true if the SQL text starts with an INSERT statement.
func IsInsert(sqlText string) bool {
	s := strings.TrimSpace(strings.ToLower(sqlText))
	return strings.HasPrefix(s, "insert ")
}

// ResolveSQLTemplateVariables replaces {{var}} placeholders with the string form of their values.
func ResolveSQLTemplateVariables(sql string, variables map[string]any) string {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return ""
	}
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}`)
	matches := re.FindAllStringSubmatch(sql, -1)
	if len(matches) == 0 {
		return sql
	}
	return re.ReplaceAllStringFunc(sql, func(s string) string {
		m := re.FindStringSubmatch(s)
		if len(m) < 2 {
			return s
		}
		return fmt.Sprint(variables[m[1]])
	})
}
