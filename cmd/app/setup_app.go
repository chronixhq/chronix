package app

import (
	"chronix/cmd/app/rpc"
	"chronix/internal/db"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/dan-sherwin/go-app-settings"
)

func Setup() {
	updater.CleanupOldExecutable()
	initLogger()
	setWorkingDir()
	dbPath := serverruntime.DataDir + "/chronix.db"
	dbExists := false
	if _, err := os.Stat(dbPath); err == nil {
		dbExists = true
	}
	slog.Debug("checking database", "path", dbPath, "exists", dbExists)

	processCLI(dbExists)

	cmd := CLICommand.Command()
	if cmd == "version show" || cmd == "version list" {
		initLogger()
		if cmd == "version list" {
			slog.Info("build info", slog.String("version", buildinfo.Version))
		}
		return
	}

	if !dbExists && cmd != "run" && !strings.HasPrefix(cmd, "version") {
		fmt.Println("Error: Application needs to be initialized first. Run application in foreground first to initialize it.")
		os.Exit(1)
	}

	_, _ = db.EnsureSQLiteFile(dbPath)
	if err := app_settings.Setup(dbPath, app_settings.SettingsOptions{
		RpcSocketPathToListRunningSettings: rpc.SocketPath,
		KongVars:                           &vars,
	}); err != nil {
		slog.Error("app_settings setup failed", "path", dbPath, "error", err)
	}
	if dbExists {
		LoggingLevel = cliConfig.AppSettings.Logging.Level
	}
	initLogger()
	slog.Info("build info", slog.String("version", buildinfo.Version))
}
