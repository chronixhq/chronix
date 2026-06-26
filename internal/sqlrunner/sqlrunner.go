package sqlrunner

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"chronix/internal/agentmux"
	"chronix/pkg/sqlutil"
	// Register pgx stdlib driver for database/sql
	_ "github.com/jackc/pgx/v5/stdlib"
	// Register mysql driver for database/sql
	_ "github.com/go-sql-driver/mysql"
	// Register mssql driver for database/sql
	_ "github.com/microsoft/go-mssqldb"
	// Register sqlite driver for database/sql
	_ "github.com/glebarez/go-sqlite"
	// Register oracle driver
	_ "github.com/sijms/go-ora/v2"
	// Register snowflake driver
	_ "github.com/snowflakedb/gosnowflake"
)

// SQLRunner defines a minimal interface to run SQL queries/execs either locally or via an agent.
type SQLRunner interface {
	Query(ctx context.Context, driver string, dsn string, sqlText string, args []any, rowCap int) (int, []map[string]any, error)
	Exec(ctx context.Context, driver string, dsn string, sqlText string, args []any) (int64, error)
	// RunInSession executes the given function with a SessionRunner that maintains session state (e.g. connection-sticky).
	RunInSession(ctx context.Context, driver string, dsn string, fn func(SessionRunner) error) error
}

// SessionRunner is a restricted SQLRunner that executes within a pre-established session/connection.
type SessionRunner interface {
	Query(ctx context.Context, sqlText string, args []any, rowCap int) (int, []map[string]any, error)
	Exec(ctx context.Context, sqlText string, args []any) (int64, error)
}

// LocalRunner executes SQL using the server's database/sql.
type LocalRunner struct{}

func (LocalRunner) open(driver, dsn string) (*sql.DB, error) {
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

func (r LocalRunner) Query(ctx context.Context, driver string, dsn string, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	db, err := r.open(driver, dsn)
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
			m := make(map[string]any, len(cols))
			for i, c := range cols {
				m[c] = vals[i]
			}
			samples = append(samples, m)
		}
		if count >= rowCap {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return count, samples, err
	}
	return count, samples, nil
}

func (r LocalRunner) Exec(ctx context.Context, driver string, dsn string, sqlText string, args []any) (int64, error) {
	db, err := r.open(driver, dsn)
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

func (r LocalRunner) RunInSession(ctx context.Context, driver string, dsn string, fn func(SessionRunner) error) error {
	db, err := r.open(driver, dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	return fn(localSessionRunner{conn: conn, driver: sqlutil.NormalizeDriver(driver)})
}

type localSessionRunner struct {
	conn   *sql.Conn
	driver string
}

func (r localSessionRunner) Query(ctx context.Context, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	rows, err := r.conn.QueryContext(ctx, sqlText, args...)
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
			m := make(map[string]any, len(cols))
			for i, c := range cols {
				m[c] = vals[i]
			}
			samples = append(samples, m)
		}
		if count >= rowCap {
			break
		}
	}
	if err := rows.Err(); err != nil {
		return count, samples, err
	}
	return count, samples, nil
}

func (r localSessionRunner) Exec(ctx context.Context, sqlText string, args []any) (int64, error) {
	res, err := r.conn.ExecContext(ctx, sqlText, args...)
	if err != nil {
		return 0, err
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return -1, nil
	}
	return ra, nil
}

// AgentRunner sends a message over an agent WebSocket and awaits the response.
type AgentRunner struct{ AgentID string }

type sqlQueryReq struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	RowCap    int    `json:"rowCap"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlQueryOK struct {
	RowsCount  int              `json:"rowsCount"`
	RowsSample []map[string]any `json:"rowsSample"`
}

type sqlExecReq struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlExecOK struct {
	RowsAffected int64 `json:"rowsAffected"`
}

type sqlError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Detail  map[string]any `json:"detail"`
}

type sqlSessionBeginReq struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionBeginOK struct {
	SessionID string `json:"sessionId"`
}

type sqlSessionQueryReq struct {
	SessionID string `json:"sessionId"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	RowCap    int    `json:"rowCap"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionExecReq struct {
	SessionID string `json:"sessionId"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionEndReq struct {
	SessionID string `json:"sessionId"`
}

func (a AgentRunner) Query(ctx context.Context, driver string, dsn string, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	c := agentmux.DefaultManager.Get(a.AgentID)
	if c == nil {
		return 0, nil, errors.New("agent not connected")
	}
	toMs := int64(time.Until(deadlineOr(ctx, 5*time.Second)).Milliseconds())
	payload := sqlQueryReq{Driver: sqlutil.NormalizeDriver(driver), DSN: dsn, SQL: sqlText, Args: args, RowCap: rowCap, TimeoutMs: toMs}
	respType, respBytes, err := c.Request(ctx, "sql.query", payload)
	if err != nil {
		return 0, nil, err
	}
	switch respType {
	case "sql.query.ok":
		var ok sqlQueryOK
		if err := json.Unmarshal(respBytes, &ok); err != nil {
			return 0, nil, err
		}
		return ok.RowsCount, ok.RowsSample, nil
	case "sql.error":
		var e sqlError
		_ = json.Unmarshal(respBytes, &e)
		if e.Message == "" {
			e.Message = "agent sql error"
		}
		return 0, nil, fmt.Errorf("%s", e.Message)
	default:
		return 0, nil, fmt.Errorf("unexpected response: %s", respType)
	}
}

func (a AgentRunner) Exec(ctx context.Context, driver string, dsn string, sqlText string, args []any) (int64, error) {
	c := agentmux.DefaultManager.Get(a.AgentID)
	if c == nil {
		return 0, errors.New("agent not connected")
	}
	toMs := int64(time.Until(deadlineOr(ctx, 5*time.Second)).Milliseconds())
	payload := sqlExecReq{Driver: sqlutil.NormalizeDriver(driver), DSN: dsn, SQL: sqlText, Args: args, TimeoutMs: toMs}
	respType, respBytes, err := c.Request(ctx, "sql.exec", payload)
	if err != nil {
		return 0, err
	}
	switch respType {
	case "sql.exec.ok":
		var ok sqlExecOK
		if err := json.Unmarshal(respBytes, &ok); err != nil {
			return 0, err
		}
		return ok.RowsAffected, nil
	case "sql.error":
		var e sqlError
		_ = json.Unmarshal(respBytes, &e)
		if e.Message == "" {
			e.Message = "agent sql error"
		}
		return 0, fmt.Errorf("%s", e.Message)
	default:
		return 0, fmt.Errorf("unexpected response: %s", respType)
	}
}

func (a AgentRunner) RunInSession(ctx context.Context, driver string, dsn string, fn func(SessionRunner) error) error {
	c := agentmux.DefaultManager.Get(a.AgentID)
	if c == nil {
		return errors.New("agent not connected")
	}

	// Begin session
	toMs := int64(time.Until(deadlineOr(ctx, 5*time.Second)).Milliseconds())
	beginPayload := sqlSessionBeginReq{Driver: sqlutil.NormalizeDriver(driver), DSN: dsn, TimeoutMs: toMs}
	respType, respBytes, err := c.Request(ctx, "sql.session.begin", beginPayload)
	if err != nil {
		return err
	}
	if respType != "sql.session.begin.ok" {
		return fmt.Errorf("failed to begin session: %s", respType)
	}
	var beginOK sqlSessionBeginOK
	if err := json.Unmarshal(respBytes, &beginOK); err != nil {
		return err
	}
	sessionID := beginOK.SessionID

	// Defer end session
	defer func() {
		endPayload := sqlSessionEndReq{SessionID: sessionID}
		_, _, _ = c.Request(context.Background(), "sql.session.end", endPayload)
	}()

	return fn(agentSessionRunner{agentID: a.AgentID, sessionID: sessionID})
}

type agentSessionRunner struct {
	agentID   string
	sessionID string
}

func (r agentSessionRunner) Query(ctx context.Context, sqlText string, args []any, rowCap int) (int, []map[string]any, error) {
	c := agentmux.DefaultManager.Get(r.agentID)
	if c == nil {
		return 0, nil, errors.New("agent not connected")
	}
	toMs := int64(time.Until(deadlineOr(ctx, 5*time.Second)).Milliseconds())
	payload := sqlSessionQueryReq{SessionID: r.sessionID, SQL: sqlText, Args: args, RowCap: rowCap, TimeoutMs: toMs}
	respType, respBytes, err := c.Request(ctx, "sql.session.query", payload)
	if err != nil {
		return 0, nil, err
	}
	switch respType {
	case "sql.query.ok":
		var ok sqlQueryOK
		if err := json.Unmarshal(respBytes, &ok); err != nil {
			return 0, nil, err
		}
		return ok.RowsCount, ok.RowsSample, nil
	case "sql.error":
		var e sqlError
		_ = json.Unmarshal(respBytes, &e)
		if e.Message == "" {
			e.Message = "agent sql error"
		}
		return 0, nil, fmt.Errorf("%s", e.Message)
	default:
		return 0, nil, fmt.Errorf("unexpected response: %s", respType)
	}
}

func (r agentSessionRunner) Exec(ctx context.Context, sqlText string, args []any) (int64, error) {
	c := agentmux.DefaultManager.Get(r.agentID)
	if c == nil {
		return 0, errors.New("agent not connected")
	}
	toMs := int64(time.Until(deadlineOr(ctx, 5*time.Second)).Milliseconds())
	payload := sqlSessionExecReq{SessionID: r.sessionID, SQL: sqlText, Args: args, TimeoutMs: toMs}
	respType, respBytes, err := c.Request(ctx, "sql.session.exec", payload)
	if err != nil {
		return 0, err
	}
	switch respType {
	case "sql.exec.ok":
		var ok sqlExecOK
		if err := json.Unmarshal(respBytes, &ok); err != nil {
			return 0, err
		}
		return ok.RowsAffected, nil
	case "sql.error":
		var e sqlError
		_ = json.Unmarshal(respBytes, &e)
		if e.Message == "" {
			e.Message = "agent sql error"
		}
		return 0, fmt.Errorf("%s", e.Message)
	default:
		return 0, fmt.Errorf("unexpected response: %s", respType)
	}
}

// Helpers

func deadlineOr(ctx context.Context, fallback time.Duration) time.Time {
	if dl, ok := ctx.Deadline(); ok {
		return dl
	}
	return time.Now().Add(fallback)
}
