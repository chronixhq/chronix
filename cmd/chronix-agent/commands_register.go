package main

import (
	regpkg "chronix-agent/agentregister"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

func registerMain() {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	serverFlag := fs.String("server", "", "Server host or host:port")
	nameFlag := fs.String("name", "", "Agent display name")
	forceFlag := fs.Bool("force", false, "Force re-registration even if already registered")

	if err := fs.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	host := *serverFlag
	name := *nameFlag
	if host == "" && fs.NArg() > 0 {
		host = fs.Arg(0)
	}
	if name == "" && fs.NArg() > 1 {
		name = fs.Arg(1)
	}

	cfgPath := defaultConfigPath()
	if cfg, err := loadConfig(cfgPath); err == nil && cfg != nil && cfg.UUID != "" {
		if !*forceFlag {
			fmt.Printf("Agent is already registered as '%s' (UUID: %s).\n", cfg.Name, cfg.UUID)
			fmt.Println("To create a new identity, run 'chronix-agent unregister' or use --force.")
			fmt.Println("Using the saved registration.")
			return
		}
		fmt.Println("Force flag detected. Creating a new agent identity...")
	}

	if host == "" || name == "" {
		printRegisterUsage()
		os.Exit(1)
	}

	serverHost, wsPort := parseServerFlag(host)

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Unable to generate agent keypair:", err)
		os.Exit(1)
	}

	cfg := &agentConfig{
		ServerHost: serverHost,
		WSPort:     wsPort,
		UUID:       uuid.New().String(),
		Name:       name,
		PrivKeyB64: base64.StdEncoding.EncodeToString(priv),
		PubKeyB64:  base64.StdEncoding.EncodeToString(pub),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()

	fmt.Printf("Registering agent '%s' (UUID: %s) with server %s:%d...\n", cfg.Name, cfg.UUID, cfg.ServerHost, cfg.WSPort)
	err = regpkg.RegisterWithServer(ctx, cfg, Version)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, regpkg.FormatError(err))
		os.Exit(regpkg.ExitCode(err))
	}

	if err := saveConfig(cfgPath, cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Registration succeeded, but saving the agent config failed:", err)
		os.Exit(1)
	}

	fmt.Println("Registration complete.")
	fmt.Printf("Saved agent configuration: %s\n", cfgPath)
	fmt.Printf("Agent connection target: wss://%s:%d/agent/connect\n", cfg.ServerHost, cfg.WSPort)
	fmt.Println()
	fmt.Println("Starting agent...")
}
