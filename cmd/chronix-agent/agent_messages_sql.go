package main

import (
	runpkg "chronix-agent/agentrun"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type sqlQueryRequest struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	RowCap    int    `json:"rowCap"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlExecRequest struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionBeginRequest struct {
	Driver    string `json:"driver"`
	DSN       string `json:"dsn"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionQueryRequest struct {
	SessionID string `json:"sessionId"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	RowCap    int    `json:"rowCap"`
	TimeoutMs int64  `json:"timeoutMs"`
}

type sqlSessionExecRequest struct {
	SessionID string `json:"sessionId"`
	SQL       string `json:"sql"`
	Args      []any  `json:"args"`
	TimeoutMs int64  `json:"timeoutMs"`
}

func handleSQLQueryMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req sqlQueryRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	sqlOp, preview := sqlOpAndPreview(req.SQL)
	argT := argTypes(req.Args)
	opTag := "sql.query"
	if strings.EqualFold(strings.TrimSpace(req.SQL), "select 1") && req.RowCap <= 1 {
		opTag = "sql.test"
	}
	slog.Info("incoming sql request", "component", "agent", "op", opTag, "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"driver", strings.ToLower(req.Driver), "dsn_masked", maskDSN(req.Driver, req.DSN), "sql_op", sqlOp, "sql_preview", preview,
		"args_count", len(req.Args), "args_types", argT, "row_cap", req.RowCap, "timeout_ms", req.TimeoutMs)

	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()
	rowsCount, rowsSample, err := runpkg.RunQuery(taskCtx, req.Driver, req.DSN, req.SQL, req.Args, req.RowCap)
	inflight.clear(env.ID)
	if err != nil {
		slog.Error("sql query failed", "component", "agent", "op", opTag, "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "sql.error", env.ID, "db_query_failed", err.Error())
		return
	}

	slog.Info("sql query ok", "component", "agent", "op", opTag, "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "rows_count", rowsCount, "duration_ms", time.Since(start).Milliseconds())
	resp := struct {
		RowsCount  int              `json:"rowsCount"`
		RowsSample []map[string]any `json:"rowsSample"`
	}{RowsCount: rowsCount, RowsSample: rowsSample}
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.query.ok", ID: env.ID, Payload: resp})
}

func handleSQLExecMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req sqlExecRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	sqlOp, preview := sqlOpAndPreview(req.SQL)
	argT := argTypes(req.Args)
	slog.Info("incoming sql exec", "component", "agent", "op", "sql.exec", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"driver", strings.ToLower(req.Driver), "dsn_masked", maskDSN(req.Driver, req.DSN), "sql_op", sqlOp, "sql_preview", preview,
		"args_count", len(req.Args), "args_types", argT, "timeout_ms", req.TimeoutMs)

	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()
	rowsAffected, err := runpkg.RunExec(taskCtx, req.Driver, req.DSN, req.SQL, req.Args)
	inflight.clear(env.ID)
	if err != nil {
		slog.Error("sql exec failed", "component", "agent", "op", "sql.exec", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "sql.error", env.ID, "db_exec_failed", err.Error())
		return
	}

	slog.Info("sql exec ok", "component", "agent", "op", "sql.exec", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "rows_affected", rowsAffected, "duration_ms", time.Since(start).Milliseconds())
	resp := struct {
		RowsAffected int64 `json:"rowsAffected"`
	}{RowsAffected: rowsAffected}
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.exec.ok", ID: env.ID, Payload: resp})
}

func handleSQLSessionBeginMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope) {
	var req sqlSessionBeginRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	sessionID, err := runpkg.BeginSQLSession(taskCtx, req.Driver, req.DSN)
	cancel()
	if err != nil {
		slog.Error("sql session begin failed", "component", "agent", "op", "sql.session.begin", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "sql.error", env.ID, "db_session_begin_failed", err.Error())
		return
	}

	slog.Info("sql session begin ok", "component", "agent", "op", "sql.session.begin", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "session_id", sessionID, "duration_ms", time.Since(start).Milliseconds())
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.session.begin.ok", ID: env.ID, Payload: struct {
		SessionID string `json:"sessionId"`
	}{SessionID: sessionID}})
}

func handleSQLSessionQueryMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req sqlSessionQueryRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	sqlOp, preview := sqlOpAndPreview(req.SQL)
	slog.Info("incoming session sql request", "component", "agent", "op", "sql.session.query", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"sql_op", sqlOp, "sql_preview", preview, "row_cap", req.RowCap, "timeout_ms", req.TimeoutMs)

	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()
	rowsCount, rowsSample, err := runpkg.QuerySQLSession(taskCtx, req.SessionID, req.SQL, req.Args, req.RowCap)
	inflight.clear(env.ID)
	if err != nil {
		slog.Error("session sql query failed", "component", "agent", "op", "sql.session.query", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "sql.error", env.ID, "db_query_failed", err.Error())
		return
	}

	slog.Info("session sql query ok", "component", "agent", "op", "sql.session.query", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID, "rows_count", rowsCount, "duration_ms", time.Since(start).Milliseconds())
	resp := struct {
		RowsCount  int              `json:"rowsCount"`
		RowsSample []map[string]any `json:"rowsSample"`
	}{RowsCount: rowsCount, RowsSample: rowsSample}
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.query.ok", ID: env.ID, Payload: resp})
}

func handleSQLSessionExecMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req sqlSessionExecRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	sqlOp, preview := sqlOpAndPreview(req.SQL)
	slog.Info("incoming session sql exec", "component", "agent", "op", "sql.session.exec", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"sql_op", sqlOp, "sql_preview", preview, "timeout_ms", req.TimeoutMs)

	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()
	rowsAffected, err := runpkg.ExecSQLSession(taskCtx, req.SessionID, req.SQL, req.Args)
	inflight.clear(env.ID)
	if err != nil {
		slog.Error("session sql exec failed", "component", "agent", "op", "sql.session.exec", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "sql.error", env.ID, "db_exec_failed", err.Error())
		return
	}

	slog.Info("session sql exec ok", "component", "agent", "op", "sql.session.exec", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID, "rows_affected", rowsAffected, "duration_ms", time.Since(start).Milliseconds())
	resp := struct {
		RowsAffected int64 `json:"rowsAffected"`
	}{RowsAffected: rowsAffected}
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.exec.ok", ID: env.ID, Payload: resp})
}

func handleSQLSessionEndMessage(conn *websocket.Conn, cfg *agentConfig, env agentEnvelope) {
	var req struct {
		SessionID string `json:"sessionId"`
	}
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "sql.error", env.ID, "bad_request", "invalid payload")
		return
	}

	slog.Info("closing sql session", "component", "agent", "op", "sql.session.end", "id", env.ID, "session_id", req.SessionID, "agent", cfg.Name, "agent_id", cfg.UUID)
	runpkg.EndSQLSession(req.SessionID)
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "sql.session.end.ok", ID: env.ID, Payload: struct{}{}})
}
