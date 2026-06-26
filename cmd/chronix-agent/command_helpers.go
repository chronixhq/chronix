package main

import (
	regpkg "chronix-agent/agentregister"
	"fmt"
	"os"
	"strings"
)

func isRegisteredConfig(cfg *agentConfig) bool {
	return cfg != nil &&
		strings.TrimSpace(cfg.UUID) != "" &&
		strings.TrimSpace(cfg.PrivKeyB64) != "" &&
		strings.TrimSpace(cfg.PubKeyB64) != "" &&
		strings.TrimSpace(cfg.ServerHost) != "" &&
		cfg.WSPort > 0
}

func loadRegisteredConfigForCommand() (string, *agentConfig) {
	cfgPath := defaultConfigPath()
	cfg, err := loadConfig(cfgPath)
	if err != nil || !isRegisteredConfig(cfg) {
		fmt.Printf("No saved agent registration found at %s\n", cfgPath)
		fmt.Println("Run: chronix-agent register <server[:port]> <name>")
		os.Exit(1)
	}
	return cfgPath, cfg
}

func currentLocalAgentStatus() (AgentStatus, bool) {
	var statusInfo AgentStatus
	if err := agentRPCCall("Agent.Status", nil, &statusInfo); err != nil {
		return AgentStatus{}, false
	}
	return statusInfo, true
}

func restartAgentIfRunning() (bool, error) {
	if _, ok := currentLocalAgentStatus(); !ok {
		return false, nil
	}
	return true, agentRPCCall("Agent.Restart", nil, nil)
}

func stopAgentIfRunning() (bool, error) {
	if _, ok := currentLocalAgentStatus(); !ok {
		return false, nil
	}
	return true, agentRPCCall("Agent.Stop", nil, nil)
}

func formatProbeError(err error) string {
	if err == nil {
		return "reachable"
	}
	msg := strings.TrimSpace(err.Error())
	if strings.Contains(strings.ToLower(msg), "server pin mismatch") {
		return "unreachable (server pin mismatch)"
	}
	return "unreachable (" + msg + ")"
}

func formatRegistrationError(prefix string, err error) string {
	msg := regpkg.FormatError(err)
	if msg == "" {
		msg = err.Error()
	}
	if prefix == "" {
		return msg
	}
	return prefix + ": " + msg
}
