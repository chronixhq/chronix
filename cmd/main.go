package main

import (
	"strings"
	"time"

	app2 "chronix/cmd/app"
	"chronix/cmd/app/consts"
	"chronix/cmd/app/rpc"
	"chronix/cxrestapi"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/svc"
	"chronix/internal/updater"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	updater.PreRestart = cxrestapi.ShutdownServer
	app2.Setup()
	processCommand()
	svc.RunIfService("chronix")
	slog.Info("Run command called. Starting " + consts.APPNAME + " as a daemon.")

	// Check if daemon is already running via RPC socket
	if cl, err := rpc.Client(); err == nil {
		_ = cl.Close()
		if svc.ServiceRunning {
			slog.Error("Chronix is already running. Service instance exiting.")
			os.Exit(1)
		}

		fmt.Print("Chronix is already running. Would you like to stop the existing instance and start this one instead? (y/N): ")
		var response string
		_, _ = fmt.Scanln(&response)
		if strings.ToLower(response) == "y" {
			if err := rpc.Call("Server.StopServer", nil, nil); err != nil {
				fmt.Printf("Failed to stop existing instance: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Existing instance stopped. Waiting 2 seconds for ports to release...")
			time.Sleep(2 * time.Second)
		} else {
			fmt.Println("Exiting.")
			return
		}
	}

	app2.SetupDaemon()
	cxrestapi.Start()
	slog.Info(consts.APPNAME + " is running.")
	if serverruntime.CurrentServerStatus == serverruntime.StatusUninitialized {
		cxrestapi.GenerateAdminCode()
		fmt.Printf("System uninitialized\nGo to %s\nAdmin login code: %s\nUse this temporary admin code to login to Chronix.\nThis code is good for 10 minutes.\n", cxrestapi.GetServerUIURL(), cxrestapi.ActiveAdminCode)
	}

	// Context and signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.OnStop = cancel
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
		sig := <-ch
		slog.Info("Shutdown signal received", slog.String("signal", sig.String()))
		cancel()
	}()

	<-ctx.Done()
	slog.Info("Shutdown initiated")

	slog.Info("Shutting down RPC server")
	rpc.Shutdown()

	slog.Info("Syncing auth keys")
	if err := cxrestapi.SyncAuthKeys(); err != nil {
		slog.Warn("sync auth keys", "error", err)
	}

	cxrestapi.ShutdownServer()
	slog.Info(consts.APPNAME + " is shut down.")
}

func processCommand() {
	cmd := app2.CLICommand.Command()
	if cmd == "run" {
		return
	}
	slog.Info("Command called", "command", cmd)
	if err := app2.CLICommand.Run(); err != nil {
		slog.Error("run command", "error", err)
	}
	os.Exit(0)
}
