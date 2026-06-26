package main

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type agentEnvelope struct {
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Payload json.RawMessage `json:"payload"`
}

type agentInflight struct {
	mu      sync.Mutex
	cancels map[string]context.CancelFunc
}

func newAgentInflight() *agentInflight {
	return &agentInflight{cancels: map[string]context.CancelFunc{}}
}

func (a *agentInflight) set(id string, cancel context.CancelFunc) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cancels[id] = cancel
}

func (a *agentInflight) clear(id string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.cancels, id)
}

func (a *agentInflight) cancel(id string) {
	a.mu.Lock()
	cancel := a.cancels[id]
	delete(a.cancels, id)
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (a *agentInflight) cancelAll() {
	a.mu.Lock()
	cancels := a.cancels
	a.cancels = map[string]context.CancelFunc{}
	a.mu.Unlock()
	for _, cancel := range cancels {
		cancel()
	}
}

func handleAgentMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	switch env.Type {
	case "sql.cancel":
		inflight.cancel(env.ID)
	case "sql.query":
		handleSQLQueryMessage(ctx, conn, cfg, env, inflight)
	case "sql.exec":
		handleSQLExecMessage(ctx, conn, cfg, env, inflight)
	case "sql.session.begin":
		handleSQLSessionBeginMessage(ctx, conn, cfg, env)
	case "sql.session.query":
		handleSQLSessionQueryMessage(ctx, conn, cfg, env, inflight)
	case "sql.session.exec":
		handleSQLSessionExecMessage(ctx, conn, cfg, env, inflight)
	case "sql.session.end":
		handleSQLSessionEndMessage(conn, cfg, env)
	case "shell.run":
		handleShellRunMessage(ctx, conn, cfg, env, inflight)
	case "webtask.run":
		handleWebTaskMessage(ctx, conn, cfg, env, inflight)
	case "agent.update":
		handleAgentUpdateMessage(conn, cfg, env)
	case "agent.restart":
		handleAgentRestartMessage(conn, env)
	}
}
