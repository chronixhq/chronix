package app

import (
	"chronix/cmd/app/rpc"
	"chronix/cmd/app/systemdata"
	cxrestapi "chronix/cxrestapi"
	"chronix/internal/connhealth"
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	eventspkg "chronix/internal/events"
	jobrunpkg "chronix/internal/jobrun"
	progresspkg "chronix/internal/progress"
	"chronix/internal/scheduler"
	"chronix/internal/secret"
	serverruntime "chronix/internal/serverruntime"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"context"
	"fmt"
	"log/slog"
)

func SetupDaemon() {
	slog.Debug("Chronix app daemon setup")
	progresspkg.SetBroadcaster(func(eventType string, data any) error {
		return eventspkg.BroadcastEvent(eventspkg.SSEEventType(eventType), data)
	})
	secret.InitMasterKey(serverruntime.DataDir)
	db.DbInit(serverruntime.DataDir + "/chronix.db")
	if err := db.SyncSchema(db.DB); err != nil {
		slog.Error("Database schema sync failed", "error", err)
	}
	db.DB.Exec("PRAGMA journal_mode=WAL;")
	db.DB.Exec("PRAGMA synchronous=NORMAL;")
	updater.DataDir = serverruntime.DataDir
	cxsettingspkg.LoadCxSettings()
	systemdata.StartSystemDataUpdates()
	systemdata.StartAgentRegistrationCleanup()
	systemdata.StartLogRetentionWorker()
	updater.StartBackgroundUpdater(buildinfo.Version)
	serverruntime.UpdateServerStatus()
	jobrunpkg.SetupJobRunner(context.Background())
	scheduler.Start(context.Background())
	connhealth.Start(context.Background())
	cxrestapi.SetupAuthKeys()
	if err := rpc.StartServer(); err != nil {
		slog.Error("rpc start server", "error", err)
	}
	slog.Debug("Setup complete")
	slog.Info(fmt.Sprintf("Data directory: %s", serverruntime.DataDir))
}
