package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/sijms/go-ora/v2"
	_ "github.com/snowflakedb/gosnowflake"
	_ "modernc.org/sqlite"
)

const (
	DefaultAgentPort = 5172
)

func main() {
	PreRestart = func() {
		stopAgentRPCServer()
	}
	cleanupOldExecutable()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--help" || arg == "-h" || arg == "help" {
			if len(os.Args) > 2 {
				sub := os.Args[2]
				switch sub {
				case "register":
					printRegisterUsage()
					return
				case "reset":
					printResetUsage()
					return
				case "unregister":
					printUnregisterUsage()
					return
				case "service":
					printServiceUsage()
					return
				case "info":
					printInfoUsage()
					return
				case "repoint":
					printRepointUsage()
					return
				case "repair-register":
					printRepairRegisterUsage()
					return
				case "rename":
					printRenameUsage()
					return
				}
			}
			printTopUsage()
			return
		}
		if arg == "version" || arg == "-v" {
			fmt.Println(Version)
			if ReleaseNotes != "" {
				fmt.Printf("\nRelease Notes:\n%s\n", ReleaseNotes)
			}
			return
		}
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "register":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printRegisterUsage()
					return
				}
			}
			registerMain()
		case "info":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printInfoUsage()
					return
				}
			}
			infoMain()
			return
		case "repoint":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printRepointUsage()
					return
				}
			}
			repointMain()
			return
		case "repair-register":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printRepairRegisterUsage()
					return
				}
			}
			repairRegisterMain()
			return
		case "rename":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printRenameUsage()
					return
				}
			}
			renameMain()
			return
		case "reset":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printResetUsage()
					return
				}
			}
			resetMain()
			return
		case "unregister":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printUnregisterUsage()
					return
				}
			}
			unregisterMain()
			return
		case "run":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printRunUsage()
					return
				}
			}
		case "version":
			for _, a := range os.Args[2:] {
				if a == "-h" || a == "--help" {
					printVersionUsage()
					return
				}
			}
			fmt.Println(Version)
			if ReleaseNotes != "" {
				fmt.Printf("\nRelease Notes:\n%s\n", ReleaseNotes)
			}
			return
		case "service":
			if len(os.Args) < 3 {
				printServiceUsage()
				return
			}
			action := os.Args[2]
			switch action {
			case "install":
				installMain()
			case "uninstall":
				uninstallMain()
			case "start", "stop", "status":
				cfg := SvcConfig{Name: "chronix-agent"}
				if err := handleAgentService(cfg, action); err != nil {
					fmt.Printf("Error: %v\n", err)
					os.Exit(1)
				}
			default:
				printServiceUsage()
			}
			return
		case "stop":
			if err := agentRPCCall("Agent.Stop", nil, nil); err != nil {
				fmt.Printf("Error stopping agent: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Stop command sent to Chronix agent.")
			return
		case "status":
			var statusInfo AgentStatus
			if err := agentRPCCall("Agent.Status", nil, &statusInfo); err != nil {
				fmt.Printf("Chronix agent is not running (error: %v).\n", err)
				os.Exit(0)
			}
			fmt.Printf("Chronix agent is %s (service: %v).\n", statusInfo.Status, statusInfo.IsService)
			return
		case "restart":
			if err := agentRPCCall("Agent.Restart", nil, nil); err != nil {
				fmt.Printf("Error restarting agent: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Restart command sent to Chronix agent.")
			return
		}
	}

	runAgentIfService("chronix-agent")
	cfgPath := defaultConfigPath()
	cfg, err := loadConfig(cfgPath)
	if err != nil || cfg == nil || cfg.UUID == "" || cfg.PrivKeyB64 == "" || cfg.PubKeyB64 == "" {
		printTopUsage()
		fmt.Println()
		fmt.Println("No agent registration found. Run: chronix-agent register <server[:port]> <name>")
		os.Exit(2)
	}

	var statusInfo AgentStatus
	if err := agentRPCCall("Agent.Status", nil, &statusInfo); err == nil {
		if agentServiceRunning {
			slog.Error("Chronix Agent is already running. Service instance exiting.")
			os.Exit(1)
		}

		if statusInfo.IsService {
			fmt.Println("Chronix Agent is already running as a service. It should be managed through systemd.")
			os.Exit(1)
		}

		fmt.Println("Chronix Agent is already running. Force stopping the existing instance...")
		if err := agentRPCCall("Agent.Stop", nil, nil); err != nil {
			fmt.Printf("Failed to stop existing instance: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Existing instance stopped. Waiting 1 second...")
		time.Sleep(1 * time.Second)
	}

	if err := startAgentRPCServer(); err != nil {
		slog.Error("Failed to start agent RPC server", "error", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	onAgentStop = cancel
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		cancel()
	}()

	runAgent(ctx, cfg)

	<-ctx.Done()
	slog.Info("Shutdown initiated")
	stopAgentRPCServer()
	slog.Info("Chronix Agent is shut down.")
}
