package main

import (
	"bufio"
	regpkg "chronix-agent/agentregister"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

func resetMain() {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	yesFlag := fs.Bool("y", false, "Skip confirmation prompt")
	_ = fs.Parse(os.Args[2:])

	cfgPath := defaultConfigPath()
	cfg, err := loadConfig(cfgPath)
	if err != nil || cfg == nil {
		fmt.Printf("No saved agent registration found at %s\n", cfgPath)
		fmt.Println("Run: chronix-agent register <server[:port]> <name>")
		return
	}

	if !*yesFlag {
		fmt.Printf("Are you sure you want to clear the saved server pin at %s?\n", cfgPath)
		fmt.Println("This will make the next connection trust and pin the server again (TOFU).")
		fmt.Print("Type 'RESET' to confirm: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) != "RESET" {
				fmt.Println("Reset cancelled.")
				return
			}
		} else {
			fmt.Println("Reset cancelled.")
			return
		}
	}

	cfg.ServerSPKIB64 = ""
	if err := saveConfig(cfgPath, cfg); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to update registration:", err)
		os.Exit(1)
	}
	fmt.Println("Saved server pin cleared.")
	fmt.Println("The next connection will trust and pin the server again.")
}

func unregisterMain() {
	fs := flag.NewFlagSet("unregister", flag.ExitOnError)
	yesFlag := fs.Bool("y", false, "Skip confirmation prompt")
	_ = fs.Parse(os.Args[2:])

	cfgPath := defaultConfigPath()
	cfg, err := loadConfig(cfgPath)
	if os.IsNotExist(err) || cfg == nil {
		fmt.Printf("No saved agent registration found at %s\n", cfgPath)
		fmt.Println("Run: chronix-agent register <server[:port]> <name>")
		return
	}

	if !*yesFlag {
		fmt.Printf("Are you sure you want to remove the local registration at %s?\n", cfgPath)
		fmt.Println("This deletes the local agent identity (UUID and keys). You will need to register again before the agent can reconnect.")
		fmt.Print("Type 'UNREGISTER' to confirm: ")
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) != "UNREGISTER" {
				fmt.Println("Unregister cancelled.")
				return
			}
		} else {
			fmt.Println("Unregister cancelled.")
			return
		}
	}

	if isRegisteredConfig(cfg) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := regpkg.UnregisterFromServer(ctx, cfg); err != nil {
			fmt.Printf("Warning: could not notify the server before removing the local registration: %v\n", err)
		} else {
			fmt.Println("Server record marked as unregistered.")
		}
	}

	if stopped, err := stopAgentIfRunning(); err != nil {
		fmt.Printf("Warning: could not stop the running agent before removing the local registration: %v\n", err)
	} else if stopped {
		fmt.Println("Local agent stopped.")
	}

	if err := os.Remove(cfgPath); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to remove registration file:", err)
		os.Exit(1)
	}
	fmt.Println("Local registration removed.")
}
