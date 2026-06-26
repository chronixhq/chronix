package main

import (
	"fmt"
	"os"
)

func installMain() {
	cfgPath := defaultConfigPath()
	cfg, err := loadConfig(cfgPath)
	if err != nil || cfg == nil || cfg.UUID == "" || cfg.PrivKeyB64 == "" || cfg.PubKeyB64 == "" {
		fmt.Println("Error: Agent must be registered before installing as a service. Try running: chronix-agent register <server[:port]> <name>")
		os.Exit(1)
	}

	svcCfg := SvcConfig{
		Name:        "chronix-agent",
		DisplayName: "Chronix Agent",
		Description: "Chronix Remote Task Agent",
		Arguments:   []string{"run"},
	}
	if err := handleAgentService(svcCfg, "install"); err != nil {
		fmt.Printf("Error installing agent service: %v\n", err)
		os.Exit(1)
	}
}

func uninstallMain() {
	cfg := SvcConfig{
		Name: "chronix-agent",
	}
	if err := handleAgentService(cfg, "uninstall"); err != nil {
		fmt.Printf("Error uninstalling agent service: %v\n", err)
		os.Exit(1)
	}
}
