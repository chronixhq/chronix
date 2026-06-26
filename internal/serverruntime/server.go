package serverruntime

import (
	"chronix/cmd/app/rpc"
	"chronix/internal/agentmux"
	chronixreactapp "chronix/internal/chronix_react_app"
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	eventspkg "chronix/internal/events"
	"chronix/internal/svc"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"io/fs"
	"net/http"
)

type (
	ServerStatus    string
	embedFileSystem struct {
		http.FileSystem
	}
)

var spaFS fs.FS

func init() {
	var err error
	spaFS, err = chronixreactapp.DistSubFS()
	if err != nil {
		slog.Error("embed dist sub fs", "error", err)
	}
}

const (
	StatusUnknown       ServerStatus = "unknown"
	StatusUninitialized ServerStatus = "uninitialized"
	StatusSuspended     ServerStatus = "suspended"
	StatusStartingUp    ServerStatus = "startingup"
	StatusActive        ServerStatus = "active"
)

var (
	CurrentServerStatus = StatusUnknown
	suspendLock         = false
)

func init() {
	CurrentServerStatus = StatusStartingUp
	if err := rpc.RegisterName("Server", &ServerCommandDef{}); err != nil {
		slog.Error("rpc register server", "error", err)
	}
}

func AnonymousServerRouter(a *gin.Engine) {
	a.Use(static.Serve("/", embedFolder(chronixreactapp.DistFS, "dist")))
	a.NoRoute(func(c *gin.Context) {
		f, _ := embedFolder(chronixreactapp.DistFS, "dist").Open("/index.html")
		http.ServeContent(c.Writer, c.Request, "/", time.Now(), f)
	})

	a.GET("/server/status", getServerStatus)

}

func getServerStatus(c *gin.Context) {
	restresponse.RestSuccess(c, gin.H{"status": CurrentServerStatus})
}

func UpdateServerStatus() {
	defer func() {
		if err := eventspkg.BroadcastEvent(eventspkg.SSEEventServerStatus, CurrentServerStatus); err != nil {
			slog.Warn("broadcast server status", "error", err)
		}
	}()
	if suspendLock {
		CurrentServerStatus = StatusSuspended
		return
	}
	if utilities.PtrVal(cxsettingspkg.CxSettings.ServerURL) == "" {
		CurrentServerStatus = StatusUninitialized
		return
	}
	_, err := db.CxUser.Where(db.CxUser.Admin.Is(true), db.CxUser.Enabled.Is(true)).Take()
	if err != nil {
		CurrentServerStatus = StatusUninitialized
		return
	}
	CurrentServerStatus = StatusActive
}

type (
	ServerCommandDef struct {
		SuspendServerDef   SuspendServerCommand   `cmd:"" help:"Suspend the server" name:"suspendServer" hidden:""`
		UnSuspendServerDef UnSuspendServerCommand `cmd:"" help:"Unsuspend the server" name:"unsuspendServer" hidden:""`
		RestartDef         RestartCommand         `cmd:"" help:"Restart the server" name:"restart"`
		StopDef            StopCommand            `cmd:"" help:"Stop the server" name:"stop"`
		StatusDef          StatusCommand          `cmd:"" help:"Check server status" name:"status"`
		Service            ServiceDef             `cmd:"" help:"Manage the native system service"`
		ListAgentsDef      ListAgentsCommand      `cmd:"" help:"List connected agents" name:"listAgents" hidden:""`
		UpdateAgentDef     UpdateAgentCommand     `cmd:"" help:"Update an agent" name:"updateAgent" hidden:""`
		CheckForUpdatesDef CheckForUpdatesCommand `cmd:"" help:"Check for updates" name:"checkForUpdates" hidden:""`
	}
	ServiceDef struct {
		Install   ServiceInstallCommand   `cmd:"" help:"Install as a native system service (macOS/Linux require sudo)" group:"Service"`
		Uninstall ServiceUninstallCommand `cmd:"" help:"Uninstall the native service (macOS/Linux require sudo)" group:"Service"`
		Start     ServiceStartCommand     `cmd:"" help:"Start the service (macOS/Linux require sudo)" group:"Service"`
		Stop      ServiceStopCommand      `cmd:"" help:"Stop the service (macOS/Linux require sudo)" group:"Service"`
		Status    ServiceStatusCommand    `cmd:"" help:"Check the service status" group:"Service"`
	}
	SuspendServerCommand    struct{}
	UnSuspendServerCommand  struct{}
	RestartCommand          struct{}
	StopCommand             struct{}
	StatusCommand           struct{}
	ServiceInstallCommand   struct{}
	ServiceUninstallCommand struct{}
	ServiceStartCommand     struct{}
	ServiceStopCommand      struct{}
	ServiceStatusCommand    struct{}
	ListAgentsCommand       struct{}
	CheckForUpdatesCommand  struct{}
	UpdateAgentCommand      struct {
		UUID    string `arg:""`
		Version string `arg:"" optional:""`
	}
)

type AgentRPCInfo struct {
	UUID            string
	Name            string
	Version         string
	Online          bool
	UpdateAvailable string
}

func (scd *ServerCommandDef) ListAgents(_ *struct{}, reply *[]AgentRPCInfo) error {
	// If manifest is nil or older than 5 minutes, refresh it
	if updater.AvailableVersion == nil || time.Since(updater.LastCheckTime) > 5*time.Minute {
		_, _, _ = updater.CheckForUpdates(buildinfo.Version)
	}

	agents, err := db.Agent.Find()
	if err != nil {
		return err
	}

	availableAgentVersion := ""
	if updater.AvailableVersion != nil {
		availableAgentVersion = updater.AvailableVersion.Agent.Version
	}

	onlineIDs := agentmux.DefaultManager.List()
	onlineMap := make(map[string]bool)
	for _, id := range onlineIDs {
		onlineMap[id] = true
	}

	for _, a := range agents {
		info := AgentRPCInfo{
			UUID:    a.UUID,
			Name:    a.Name,
			Version: utilities.PtrVal(a.Version),
			Online:  onlineMap[a.UUID],
		}
		if availableAgentVersion != "" && info.Version != availableAgentVersion {
			info.UpdateAvailable = availableAgentVersion
		}
		*reply = append(*reply, info)
	}

	return nil
}

func (scd *ServerCommandDef) CheckForUpdates(_ *struct{}, reply *updater.VersionManifest) error {
	manifest, _, err := updater.CheckForUpdates(buildinfo.Version)
	if err != nil {
		return err
	}
	*reply = *manifest
	return nil
}

func (scd *ServerCommandDef) UpdateAgent(args *UpdateAgentCommand, _ *struct{}) error {
	ids := []string{}
	if args.UUID == "all" {
		ids = agentmux.DefaultManager.List()
	} else {
		ids = append(ids, args.UUID)
	}

	version := args.Version
	if version == "" {
		if updater.AvailableVersion == nil || time.Since(updater.LastCheckTime) > 5*time.Minute {
			_, _, _ = updater.CheckForUpdates(buildinfo.Version)
		}
		if updater.AvailableVersion != nil {
			version = updater.AvailableVersion.Agent.Version
		}
	}

	if version == "" {
		return fmt.Errorf("no version available to update to")
	}

	for _, id := range ids {
		conn := agentmux.DefaultManager.Get(id)
		if conn == nil {
			continue
		}
		go func(c *agentmux.Conn) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			payload := map[string]string{"version": version}
			_, _, err := c.Request(ctx, "agent.update", payload)
			if err != nil {
				slog.Error("failed to trigger agent update via RPC", "id", c.ID(), "error", err)
			}
		}(conn)
	}

	return nil
}

func (*SuspendServerCommand) Run() error {
	if err := rpc.Call("Server.SetSuspendServer", true, nil); err != nil {
		slog.Error("rpc call SetSuspendServer true", "error", err)
	}
	return nil
}

func (*UnSuspendServerCommand) Run() error {
	if err := rpc.Call("Server.SetSuspendServer", false, nil); err != nil {
		slog.Error("rpc call SetSuspendServer false", "error", err)
	}
	return nil
}

func (scd *ServerCommandDef) SetSuspendServer(state *bool, _ *struct{}) error {
	suspendLock = *state
	UpdateServerStatus()
	slog.Info("Server suspended", "state", suspendLock)
	slog.Info("Server status", "status", CurrentServerStatus)
	return nil
}

func (*RestartCommand) Run() error {
	if err := rpc.Call("Server.RestartServer", nil, nil); err != nil {
		slog.Error("rpc call RestartServer", "error", err)
	}
	return nil
}

func (scd *ServerCommandDef) RestartServer(_ *struct{}, _ *struct{}) error {
	slog.Info("Restarting server via RPC")
	return updater.Restart("")
}

func (scd *ServerCommandDef) StopServer(_ *struct{}, _ *struct{}) error {
	slog.Info("Stopping server via RPC")
	go func() {
		time.Sleep(500 * time.Millisecond)
		if svc.OnStop != nil {
			svc.OnStop()
		}
		// Force exit if graceful shutdown hangs
		time.Sleep(3 * time.Second)
		slog.Warn("Graceful shutdown timed out, forcing exit")
		os.Exit(0)
	}()
	return nil
}

func (scd *ServerCommandDef) GetStatus(_ *struct{}, reply *ServerStatus) error {
	*reply = CurrentServerStatus
	return nil
}

func (*StopCommand) Run() error {
	if err := rpc.Call("Server.StopServer", nil, nil); err != nil {
		return fmt.Errorf("failed to stop server: %w", err)
	}
	fmt.Println("Stop command sent to Chronix server.")
	return nil
}

func (*StatusCommand) Run() error {
	var status ServerStatus
	if err := rpc.Call("Server.GetStatus", nil, &status); err != nil {
		fmt.Printf("Chronix server is not running (error: %v).\n", err)
		return nil
	}
	fmt.Printf("Chronix server is running.\nStatus: %s\nVersion: %s\n", status, buildinfo.Version)
	return nil
}

func (*ServiceInstallCommand) Run() error {
	cfg := svc.Config{
		Name:        "chronix",
		DisplayName: "Chronix Task Automation",
		Description: "Chronix Task Automation Server",
		Arguments:   []string{"run"},
	}
	return svc.Handle(cfg, "install")
}

func (*ServiceUninstallCommand) Run() error {
	cfg := svc.Config{
		Name: "chronix",
	}
	return svc.Handle(cfg, "uninstall")
}

func (*ServiceStartCommand) Run() error {
	cfg := svc.Config{
		Name: "chronix",
	}
	return svc.Handle(cfg, "start")
}

func (*ServiceStopCommand) Run() error {
	cfg := svc.Config{
		Name: "chronix",
	}
	return svc.Handle(cfg, "stop")
}

func (*ServiceStatusCommand) Run() error {
	cfg := svc.Config{
		Name: "chronix",
	}
	return svc.Handle(cfg, "status")
}

//goland:noinspection GoUnusedParameter
func (e embedFileSystem) Exists(_ string, path string) bool {
	_, err := e.Open(path)
	return err == nil
}

func embedFolder(fsEmbed embed.FS, targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}
