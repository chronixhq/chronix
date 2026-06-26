package main

import (
	runpkg "chronix-agent/agentrun"
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

type webTaskRunRequest struct {
	Method    string            `json:"method"`
	URL       string            `json:"url"`
	Headers   map[string]string `json:"headers"`
	Body      []byte            `json:"body"`
	TimeoutMs int64             `json:"timeoutMs"`
}

func handleWebTaskMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req webTaskRunRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "webtask.error", env.ID, "bad_request", "invalid payload")
		return
	}

	slog.Info("incoming webtask request", "component", "agent", "op", "webtask.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"method", req.Method, "url", req.URL, "timeout_ms", req.TimeoutMs)
	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()

	res, err := runpkg.RunWebTask(taskCtx, req.Method, req.URL, req.Headers, req.Body)
	inflight.clear(env.ID)
	if err != nil {
		slog.Error("webtask run failed", "component", "agent", "op", "webtask.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "webtask.error", env.ID, "webtask_run_failed", err.Error())
		return
	}

	slog.Info("webtask run complete", "component", "agent", "op", "webtask.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "status", res.StatusCode, "duration_ms", time.Since(start).Milliseconds())
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "webtask.run.ok", ID: env.ID, Payload: res})
}
