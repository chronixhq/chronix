package agentrun

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"sync"

	"chronix/pkg/sqlutil"

	"github.com/google/uuid"
)

type sqlSession struct {
	db     *sql.DB
	conn   *sql.Conn
	driver string
}

var (
	sqlSessions   = make(map[string]*sqlSession)
	sqlSessionsMu sync.Mutex
)

func openDB(driver, dsn string) (*sql.DB, error) {
	drv := sqlutil.NormalizeDriver(driver)
	dsn = sqlutil.NormalizeDSN(drv, dsn)
	switch drv {
	case "postgres", "pgx":
		return sql.Open("pgx", dsn)
	case "mysql":
		return sql.Open("mysql", dsn)
	case "mssql", "sqlserver":
		return sql.Open("sqlserver", dsn)
	case "sqlite":
		if dsn != ":memory:" && !strings.Contains(dsn, "mode=memory") {
			path := dsn
			if strings.HasPrefix(path, "file:") {
				path = strings.TrimPrefix(path, "file:")
				if idx := strings.Index(path, "?"); idx != -1 {
					path = path[:idx]
				}
			}
			path = strings.TrimPrefix(path, "//")
			if path != "" && path != ":memory:" {
				if _, err := os.Stat(path); os.IsNotExist(err) {
					return nil, fmt.Errorf("sqlite database file not found: %s", path)
				}
			}
		}
		return sql.Open("sqlite", dsn)
	case "oracle":
		return sql.Open("oracle", dsn)
	case "snowflake":
		return sql.Open("snowflake", dsn)
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
}

func RunQuery(ctx context.Context, driver, dsn, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	db, err := openDB(driver, dsn)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = db.Close() }()
	rows, err := db.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return 0, nil, err
	}
	count := 0
	samples := make([]map[string]any, 0, min(rowCap, 8))
	for rows.Next() {
		count++
		if len(samples) < rowCap {
			vals := make([]any, len(cols))
			scans := make([]any, len(cols))
			for i := range vals {
				scans[i] = &vals[i]
			}
			if err := rows.Scan(scans...); err != nil {
				return count, samples, err
			}
			row := make(map[string]any, len(cols))
			for i, col := range cols {
				row[col] = vals[i]
			}
			samples = append(samples, row)
		}
	}
	return count, samples, rows.Err()
}

func RunExec(ctx context.Context, driver, dsn, sqlText string, args []any) (int64, error) {
	db, err := openDB(driver, dsn)
	if err != nil {
		return 0, err
	}
	defer func() { _ = db.Close() }()
	res, err := db.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return -1, nil
	}
	return ra, nil
}

func BeginSQLSession(ctx context.Context, driver, dsn string) (string, error) {
	db, err := openDB(driver, dsn)
	if err != nil {
		return "", err
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return "", err
	}
	sessionID := uuid.NewString()
	sqlSessionsMu.Lock()
	sqlSessions[sessionID] = &sqlSession{db: db, conn: conn, driver: sqlutil.NormalizeDriver(driver)}
	sqlSessionsMu.Unlock()
	return sessionID, nil
}

func QuerySQLSession(ctx context.Context, sessionID string, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	sqlSessionsMu.Lock()
	sess, ok := sqlSessions[sessionID]
	sqlSessionsMu.Unlock()
	if !ok {
		return 0, nil, fmt.Errorf("session not found: %s", sessionID)
	}

	rows, err := sess.conn.QueryContext(ctx, sqlText, args...)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		return 0, nil, err
	}
	count := 0
	samples := make([]map[string]any, 0, min(rowCap, 8))
	for rows.Next() {
		count++
		if len(samples) < rowCap {
			vals := make([]any, len(cols))
			scans := make([]any, len(cols))
			for i := range vals {
				scans[i] = &vals[i]
			}
			if err := rows.Scan(scans...); err != nil {
				return count, samples, err
			}
			row := make(map[string]any, len(cols))
			for i, col := range cols {
				row[col] = vals[i]
			}
			samples = append(samples, row)
		}
	}
	return count, samples, rows.Err()
}

func ExecSQLSession(ctx context.Context, sessionID string, sqlText string, args []any) (int64, error) {
	sqlSessionsMu.Lock()
	sess, ok := sqlSessions[sessionID]
	sqlSessionsMu.Unlock()
	if !ok {
		return 0, fmt.Errorf("session not found: %s", sessionID)
	}

	res, err := sess.conn.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return -1, nil
	}
	return ra, nil
}

func EndSQLSession(sessionID string) {
	sqlSessionsMu.Lock()
	sess, ok := sqlSessions[sessionID]
	if ok {
		delete(sqlSessions, sessionID)
	}
	sqlSessionsMu.Unlock()

	if ok {
		_ = sess.conn.Close()
		_ = sess.db.Close()
	}
}
