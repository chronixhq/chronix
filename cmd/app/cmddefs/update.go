package cmddefs

import (
	"chronix/cmd/app/rpc"
	"chronix/internal/activity"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"fmt"
)

type UpdateCommandDef struct {
	Update UpdateCommand `cmd:"" help:"Manage application updates"`
}

type UpdateCommand struct {
	Check UpdateCheckCommand `cmd:"" help:"Check for updates" group:"Update"`
	Apply UpdateApplyCommand `cmd:"" help:"Apply available update" group:"Update"`
}

type UpdateCheckCommand struct{}

func (c *UpdateCheckCommand) Run() error {
	var manifest *updater.VersionManifest
	var available bool
	var err error

	var daemonManifest updater.VersionManifest
	if rpcErr := rpc.Call("Server.CheckForUpdates", nil, &daemonManifest); rpcErr == nil {
		manifest = &daemonManifest
		available = updater.IsNewerVersion(buildinfo.Version, manifest.Server.Version)
	} else {
		manifest, available, err = updater.CheckForUpdates(buildinfo.Version)
		if err != nil {
			return fmt.Errorf("check for updates: %w", err)
		}
	}

	fmt.Printf("Current version: %s\n", buildinfo.Version)
	if available {
		fmt.Printf("Update available: %s (Released: %s)\n", manifest.Server.Version, manifest.Server.ReleaseDate)
		fmt.Printf("\nRelease notes:\n%s\n", manifest.Server.ReleaseNotes)
		fmt.Printf("\nRun '%s update apply' to update now.\n", "chronix")
	} else {
		fmt.Println("Chronix is up to date.")
	}
	return nil
}

type UpdateApplyCommand struct{}

func (c *UpdateApplyCommand) Run() error {
	var manifest *updater.VersionManifest
	var available bool
	var err error

	var daemonManifest updater.VersionManifest
	if rpcErr := rpc.Call("Server.CheckForUpdates", nil, &daemonManifest); rpcErr == nil {
		manifest = &daemonManifest
		available = updater.IsNewerVersion(buildinfo.Version, manifest.Server.Version)
	} else {
		manifest, available, err = updater.CheckForUpdates(buildinfo.Version)
		if err != nil {
			return fmt.Errorf("check for updates: %w", err)
		}
	}

	if !available {
		fmt.Println("No update available.")
		return nil
	}

	version := manifest.Server.Version
	fmt.Printf("Applying update to version %s...\n", version)
	if err := updater.ApplyUpdate(manifest, false, buildinfo.Version); err != nil {
		return fmt.Errorf("apply update: %w", err)
	}

	_ = activity.RecordUserActivity(0, "Server Update Applied (CLI)", "Updated to version "+version, "", "")

	if err := rpc.Call("Server.RestartServer", nil, nil); err == nil {
		fmt.Println("Chronix update complete. Restarting Chronix daemon...")
	} else {
		fmt.Println("Chronix update complete.")
	}

	return nil
}
