package main

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

func handleAgentUpdateMessage(conn *websocket.Conn, cfg *agentConfig, env agentEnvelope) {
	var req struct {
		Version   string `json:"version"`
		SHA256    string `json:"sha256"`
		Signature string `json:"signature"`
	}
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "agent.error", env.ID, "bad_request", "invalid payload")
		return
	}

	slog.Info("incoming update request", "component", "agent", "op", "update", "id", env.ID, "version", req.Version)
	go func() {
		err := agentDownloadAndApplyUpdate(cfg, req.Version, req.SHA256, req.Signature)
		if err != nil {
			slog.Error("update failed", "component", "agent", "op", "update", "id", env.ID, "error", err)
			sendAgentError(conn, "agent.error", env.ID, "update_failed", err.Error())
		}
	}()
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "agent.update.ok", ID: env.ID, Payload: map[string]any{"status": "updating"}})
}

func handleAgentRestartMessage(conn *websocket.Conn, env agentEnvelope) {
	slog.Info("incoming restart request", "component", "agent", "op", "restart", "id", env.ID)
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "agent.restart.ok", ID: env.ID, Payload: map[string]any{"status": "restarting"}})

	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := restart(""); err != nil {
			slog.Error("restart failed", "component", "agent", "op", "restart", "error", err)
		}
	}()
}
