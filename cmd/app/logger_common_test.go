package app

import (
	"io"
	"log/slog"
	"testing"
)

func TestParseLevel_Mapping(t *testing.T) {
	cases := map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"":      slog.LevelInfo,
		"nope":  slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Fatalf("parseLevel(%q)=%v want %v", in, got, want)
		}
	}
}

func TestSetDefaultLogger_DoesNotPanic(_ *testing.T) {
	// Ensure it sets slog default without panicking
	l := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	setDefaultLogger(l)
	// Write a log via default to make sure it's usable
	slog.Info("test log", "component", "logger_common_test")
}
