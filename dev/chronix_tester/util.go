package main

import (
	"context"
	"os/signal"
	"syscall"
)

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
}

func resetState() error {
	paths, err := resolvePaths(CLI.DataDir)
	if err != nil {
		return err
	}
	return ResetStore(paths)
}
