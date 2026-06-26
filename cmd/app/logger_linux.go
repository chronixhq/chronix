package app

import (
	"log/slog"
)

var LoggingLevel = "info"

func initLogger() {
	lvl := parseLevel(LoggingLevel)
	journaldMinLevel = lvl
	logger := slog.New(&JournaldHandler{})
	setDefaultLogger(logger)
}
