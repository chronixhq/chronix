package main

import (
	regpkg "chronix-agent/agentregister"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func infoMain() {
	cfgPath, cfg := loadRegisteredConfigForCommand()

	statusInfo, running := currentLocalAgentStatus()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	probe, err := regpkg.ProbeServer(ctx, cfg)

	fmt.Println("Chronix Agent Registration")
	fmt.Printf("Name: %s\n", cfg.Name)
	fmt.Printf("UUID: %s\n", cfg.UUID)
	fmt.Printf("Server: %s:%d\n", cfg.ServerHost, cfg.WSPort)
	fmt.Printf("Config File: %s\n", cfgPath)
	if strings.TrimSpace(cfg.ServerSPKIB64) == "" {
		fmt.Println("Server Pin: not saved")
	} else {
		fmt.Println("Server Pin: saved")
	}
	if running {
		if statusInfo.IsService {
			fmt.Println("Local Agent: running as a service")
		} else {
			fmt.Println("Local Agent: running")
		}
	} else {
		fmt.Println("Local Agent: not running")
	}
	if err != nil {
		fmt.Printf("Server Probe: %s\n", formatProbeError(err))
		return
	}
	if probe != nil && probe.Reachable {
		reachability := "reachable"
		if !probe.PinVerified {
			reachability += " (server pin not saved yet)"
		}
		fmt.Printf("Server Probe: %s\n", reachability)
		return
	}
	fmt.Println("Server Probe: unknown")
}

func repointMain() {
	fs := flag.NewFlagSet("repoint", flag.ExitOnError)
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		printRepointUsage()
		os.Exit(1)
	}

	cfgPath, cfg := loadRegisteredConfigForCommand()
	oldTarget := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.WSPort)
	cfg.ServerHost, cfg.WSPort = parseServerFlag(fs.Arg(0))

	if err := saveConfig(cfgPath, cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to update registration:", err)
		os.Exit(1)
	}

	fmt.Printf("Saved new server target: %s -> %s:%d\n", oldTarget, cfg.ServerHost, cfg.WSPort)
	if restarted, err := restartAgentIfRunning(); err != nil {
		fmt.Printf("Server target was updated, but restarting the running agent failed: %v\n", err)
		os.Exit(1)
	} else if restarted {
		fmt.Println("Running agent restarted to use the new server target.")
	} else {
		fmt.Println("Server target updated. Start the agent to connect to the new server.")
	}
}

func repairRegisterMain() {
	fs := flag.NewFlagSet("repair-register", flag.ExitOnError)
	serverFlag := fs.String("server", "", "Server host or host:port")
	nameFlag := fs.String("name", "", "Agent display name")
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	performRepairRegister(strings.TrimSpace(*serverFlag), strings.TrimSpace(*nameFlag))
}

func renameMain() {
	fs := flag.NewFlagSet("rename", flag.ExitOnError)
	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}
	if fs.NArg() != 1 {
		printRenameUsage()
		os.Exit(1)
	}
	newName := strings.TrimSpace(fs.Arg(0))
	if newName == "" {
		printRenameUsage()
		os.Exit(1)
	}
	performRepairRegister("", newName)
}

func performRepairRegister(serverInput string, nameInput string) {
	cfgPath, cfg := loadRegisteredConfigForCommand()
	if serverInput != "" {
		cfg.ServerHost, cfg.WSPort = parseServerFlag(serverInput)
	}
	if nameInput != "" {
		cfg.Name = nameInput
	}
	if strings.TrimSpace(cfg.Name) == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Agent name cannot be empty.")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	fmt.Printf("Refreshing registration for agent '%s' (UUID: %s) against %s:%d using the existing identity...\n", cfg.Name, cfg.UUID, cfg.ServerHost, cfg.WSPort)
	res, err := regpkg.RepairRegistration(ctx, cfg, Version)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, formatRegistrationError("Repair registration failed", err))
		os.Exit(regpkg.ExitCode(err))
	}

	if err := saveConfig(cfgPath, cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Registration was refreshed, but saving the local agent config failed:", err)
		os.Exit(1)
	}

	switch res.Mode {
	case regpkg.RepairModeAuthenticated:
		fmt.Println("Server refreshed the existing agent record.")
	case regpkg.RepairModeApproval:
		fmt.Println("Server approved the existing agent identity through the recovery flow.")
	}

	if restarted, err := restartAgentIfRunning(); err != nil {
		fmt.Printf("Registration was refreshed, but restarting the running agent failed: %v\n", err)
		os.Exit(1)
	} else if restarted {
		fmt.Println("Running agent restarted to use the refreshed registration.")
	} else {
		fmt.Println("Registration refreshed. Start the agent when you're ready.")
	}
}
