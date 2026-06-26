package app

import (
	"chronix/cmd/app/consts"
	"chronix/pkg/buildinfo"
	"log/slog"
	"os"
	user2 "os/user"
)

func parseLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func setDefaultLogger(base *slog.Logger) {
	user, _ := user2.Current()
	child := base.With(
		slog.Int("pid", os.Getpid()),
		slog.String("user", user.Username),
		slog.String("app", consts.APPNAME),
		slog.String("version", buildinfo.Version),
	)
	slog.SetDefault(child)
}
