package app

import (
	"log/slog"
	"os"
)

var LoggingLevel = "info"

func initLogger() {
	lvl := parseLevel(LoggingLevel)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl})
	logger := slog.New(handler)
	setDefaultLogger(logger)
}
