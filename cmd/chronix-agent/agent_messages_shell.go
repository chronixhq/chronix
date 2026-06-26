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

type shellRunRequest struct {
	ShellPath  string            `json:"shellPath"`
	RunMode    string            `json:"runMode"`
	Command    *string           `json:"command"`
	ScriptText *string           `json:"scriptText"`
	WorkingDir *string           `json:"workingDir"`
	Env        map[string]string `json:"env"`
	TimeoutMs  int64             `json:"timeoutMs"`

	Sudo         bool    `json:"sudo"`
	RunAsUser    *string `json:"runAsUser"`
	SudoPassword *string `json:"sudoPassword"`

	Mode          string  `json:"mode"`
	Host          *string `json:"host"`
	Port          *int    `json:"port"`
	SSHUsername   *string `json:"sshUsername"`
	AuthMethod    *string `json:"authMethod"`
	SSHPassword   *string `json:"sshPassword"`
	SSHPrivateKey *string `json:"sshPrivateKey"`
	SSHKeyPass    *string `json:"sshKeyPass"`
}

func handleShellRunMessage(ctx context.Context, conn *websocket.Conn, cfg *agentConfig, env agentEnvelope, inflight *agentInflight) {
	var req shellRunRequest
	if err := json.Unmarshal(env.Payload, &req); err != nil {
		sendAgentError(conn, "shell.error", env.ID, "bad_request", "invalid payload")
		return
	}

	slog.Info("incoming shell request", "component", "agent", "op", "shell.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID,
		"shell", req.ShellPath, "mode", req.RunMode, "working_dir", safeStr(req.WorkingDir), "timeout_ms", req.TimeoutMs, "conn_mode", req.Mode)
	start := time.Now()
	taskCtx, cancel := context.WithTimeout(ctx, time.Duration(req.TimeoutMs)*time.Millisecond)
	inflight.set(env.ID, cancel)
	defer cancel()

	var exitCode int
	var stdout, stderr []byte
	var err error

	if strings.ToLower(req.Mode) == "ssh" {
		sshCfg := &runpkg.SSHConfig{
			Host:          safeStr(req.Host),
			Port:          safeInt(req.Port, 22),
			Username:      safeStr(req.SSHUsername),
			AuthMethod:    safeStr(req.AuthMethod),
			Password:      safeStr(req.SSHPassword),
			PrivateKey:    safeStr(req.SSHPrivateKey),
			KeyPassphrase: safeStr(req.SSHKeyPass),
			SudoPassword:  safeStr(req.SudoPassword),
		}
		exitCode, stdout, stderr, err = runpkg.RunSSH(taskCtx, sshCfg, req.ShellPath, req.RunMode, req.Command, req.ScriptText, req.WorkingDir, req.Env, req.Sudo, req.RunAsUser)
	} else {
		exitCode, stdout, stderr, err = runpkg.RunShell(taskCtx, req.ShellPath, req.RunMode, req.Command, req.ScriptText, req.WorkingDir, req.Env, req.Sudo, req.RunAsUser, req.SudoPassword)
	}

	inflight.clear(env.ID)
	if err != nil && exitCode == -1 {
		slog.Error("shell run failed", "component", "agent", "op", "shell.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "error", err, "duration_ms", time.Since(start).Milliseconds())
		sendAgentError(conn, "shell.error", env.ID, "shell_run_failed", err.Error())
		return
	}

	slog.Info("shell run complete", "component", "agent", "op", "shell.run", "id", env.ID, "agent", cfg.Name, "agent_id", cfg.UUID, "exit_code", exitCode, "duration_ms", time.Since(start).Milliseconds())
	resp := struct {
		ExitCode int    `json:"exitCode"`
		Stdout   string `json:"stdout"`
		Stderr   string `json:"stderr"`
	}{ExitCode: exitCode, Stdout: string(stdout), Stderr: string(stderr)}
	writeJSON(conn, struct {
		Type, ID string
		Payload  any
	}{Type: "shell.run.ok", ID: env.ID, Payload: resp})
}
