package main

import (
	"fmt"
	"log/slog"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"
)

var (
	agentRPCListener   net.Listener
	agentRPCSocketPath string
)

func defaultAgentRPCSocketPath() string {
	if p := os.Getenv("CHRONIX_AGENT_RPC_SOCKET_PATH"); p != "" {
		return p
	}
	configPath := defaultConfigPath()
	dir := filepath.Dir(configPath)
	return filepath.Join(dir, "agent-rpc.sock")
}

func startAgentRPCServer() error {
	agentRPCSocketPath = defaultAgentRPCSocketPath()
	_ = os.Remove(agentRPCSocketPath)
	if err := os.MkdirAll(filepath.Dir(agentRPCSocketPath), 0770); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	if err := rpc.RegisterName("Agent", &AgentRPCServer{}); err != nil {
		return err
	}

	var err error
	agentRPCListener, err = net.Listen("unix", agentRPCSocketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	slog.Info("Agent RPC server listening", "path", agentRPCSocketPath)
	_ = os.Chmod(agentRPCSocketPath, 0660)

	go func() {
		for {
			conn, err := agentRPCListener.Accept()
			if err != nil {
				return
			}
			go rpc.ServeConn(conn)
		}
	}()
	return nil
}

func stopAgentRPCServer() {
	if agentRPCListener != nil {
		_ = agentRPCListener.Close()
	}
	if agentRPCSocketPath != "" {
		_ = os.Remove(agentRPCSocketPath)
	}
}

type AgentRPCServer struct{}

func (s *AgentRPCServer) Stop(_ *struct{}, _ *struct{}) error {
	slog.Info("Stopping agent via RPC")
	go func() {
		time.Sleep(500 * time.Millisecond)
		if onAgentStop != nil {
			onAgentStop()
		}
		// Force exit if graceful shutdown hangs
		time.Sleep(3 * time.Second)
		slog.Warn("Graceful shutdown timed out, forcing exit")
		os.Exit(0)
	}()
	return nil
}

type AgentStatus struct {
	Status    string
	IsService bool
}

func (s *AgentRPCServer) Status(_ *struct{}, reply *AgentStatus) error {
	reply.Status = "running"
	reply.IsService = agentServiceRunning
	return nil
}

func (s *AgentRPCServer) Restart(_ *struct{}, _ *struct{}) error {
	slog.Info("Restarting agent via RPC")
	go func() {
		time.Sleep(500 * time.Millisecond)
		_ = restart("")
	}()
	return nil
}

func agentRPCCall(method string, args any, reply any) error {
	if args == nil {
		args = &struct{}{}
	}
	if reply == nil {
		reply = &struct{}{}
	}
	socketPath := defaultAgentRPCSocketPath()
	client, err := rpc.Dial("unix", socketPath)
	if err != nil {
		return fmt.Errorf("agent daemon not found (could not connect to %s): %w", socketPath, err)
	}
	defer func() { _ = client.Close() }()
	return client.Call(method, args, reply)
}
