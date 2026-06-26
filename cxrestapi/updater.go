package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/agentmux"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"context"
	"log/slog"
	"os"
	"time"

	"chronix/internal/db"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func init() {
	updater.TriggerAgentUpdate = func(version string) {
		slog.Info("Triggering updates for connected agents needing it", "version", version)
		ids := agentmux.DefaultManager.List()
		if len(ids) == 0 {
			return
		}

		// Fetch connected agents from DB
		agents, err := db.Agent.Where(db.Agent.UUID.In(ids...)).Find()
		if err != nil {
			slog.Error("failed to list agents for auto-update", "error", err)
			return
		}

		for _, a := range agents {
			if a.Version != nil && *a.Version == version {
				continue
			}
			agentID := a.UUID
			go func(agentID string) {
				conn := agentmux.DefaultManager.Get(agentID)
				if conn == nil {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				payload := map[string]string{
					"version": version,
				}

				_, _, err := conn.Request(ctx, "agent.update", payload)
				if err != nil {
					slog.Error("failed to trigger agent update", "id", agentID, "error", err)
				}
			}(agentID)
		}
	}

	updater.RecordActivityHook = func(action, details string) {
		_ = activitypkg.RecordUserActivity(0, action, details, "", "System")
	}
}

func updaterRouter(utApp *gin.Engine) {
	utApp.GET("/settings/updater/status", adminFunc(getUpdaterStatus))
	utApp.POST("/settings/updater/check", adminFunc(postUpdaterCheck))
	utApp.POST("/settings/updater/apply", adminFunc(postUpdaterApply))
}

func agentUpdaterRouter(app *gin.Engine) {
	// Agent relay update endpoint
	app.GET("/agent/update/:version/:platform", AgentAuthMiddleware(), getAgentUpdateBinary)
}

func getAgentUpdateBinary(c *gin.Context) {
	version := c.Param("version")
	platform := c.Param("platform")
	if version == "" || platform == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "version and platform are required")
		return
	}

	path := updater.GetAgentBinaryPath(version, platform)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		var manifest *updater.VersionManifest
		// Try to use the currently cached latest manifest if it matches
		if updater.AvailableVersion != nil && updater.AvailableVersion.Agent.Version == version {
			manifest = updater.AvailableVersion
		} else {
			// Fallback: try to get the manifest for this specific version from the versions list
			var err error
			manifest, err = updater.GetVersionManifestForRevert("", version)
			if err != nil {
				slog.Warn("Agent version manifest not found", "version", version, "error", err)
				restresponse.RestErrorRespond(c, restresponse.NotFound, "Agent binary not found for this version/platform")
				return
			}
		}

		// Ensure it's cached
		var err error
		path, err = updater.EnsureAgentBinaryCached(manifest, platform)
		if err != nil {
			slog.Error("Error caching agent binary", "version", version, "platform", platform, "error", err)
			restresponse.RestErrorRespond(c, restresponse.Internal, "Error caching agent binary", err.Error())
			return
		}
	}

	c.File(path)
}

func getUpdaterStatus(c *gin.Context) {
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// If manifest is nil or older than 5 minutes, refresh it
	if updater.AvailableVersion == nil || time.Since(updater.LastCheckTime) > 5*time.Minute {
		_, _, _ = updater.CheckForUpdates(buildinfo.Version)
	}

	restresponse.RestSuccess(c, gin.H{
		"currentVersion":   buildinfo.Version,
		"availableVersion": updater.AvailableVersion,
		"lastCheckTime":    updater.LastCheckTime,
		"enabled":          updater.Enabled,
		"mode":             updater.Mode,
		"windowStart":      updater.WindowStart,

		"agentEnabled":     updater.AgentEnabled,
		"agentMode":        updater.AgentMode,
		"agentWindowStart": updater.AgentWindowStart,
	})
}

func postUpdaterCheck(c *gin.Context) {
	manifest, available, err := updater.CheckForUpdates(buildinfo.Version)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error checking for updates", err.Error())
		return
	}
	restresponse.RestSuccess(c, gin.H{
		"available": available,
		"manifest":  manifest,
	})
}

func postUpdaterApply(c *gin.Context) {
	if updater.AvailableVersion == nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "No update available to apply")
		return
	}
	user := userFromGinContext(c)
	version := updater.AvailableVersion.Server.Version

	slog.Info("Update initiated via UI", "version", version, "user", user.Email)
	_ = activitypkg.RecordUserActivity(user.ID, "Server Update Initiated", "Updated to version "+version, c.ClientIP(), c.Request.UserAgent())

	go func() {
		// ApplyUpdate will exit the process on success
		err := updater.ApplyUpdate(updater.AvailableVersion, true, buildinfo.Version)
		if err != nil {
			slog.Error("Failed to apply update via API", "error", err)
		}
	}()
	restresponse.RestSuccess(c, gin.H{"message": "Update initiated"})
}
