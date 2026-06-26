package cmddefs

import (
	"chronix/cmd/app/consts"
	"chronix/cmd/app/rpc"
	"chronix/internal/activity"
	"chronix/internal/updater"
	"chronix/pkg/buildinfo"
	"fmt"
	"runtime/debug"
	"strings"
)

type (
	BuildInfoCommandDef struct {
		Buildinfo BuildInfoCommand `cmd:"" hidden:"" help:"Show the build information"`
	}
	BuildInfoCommand struct{}

	VersionCommandDef struct {
		Version VersionCommand `cmd:"" help:"Print version information and quit"`
	}
	VersionCommand struct {
		List   VersionListCommand   `cmd:"" help:"List all available versions" hidden:""`
		Revert VersionRevertCommand `cmd:"" help:"Revert to a specific version" hidden:""`
		Show   VersionShowCommand   `cmd:"" help:"Show version information" default:"1" hidden:""`
	}
	VersionShowCommand   struct{}
	VersionListCommand   struct{}
	VersionRevertCommand struct {
		Version string `arg:"" help:"Version to revert to"`
	}
)

func (v *VersionShowCommand) Run() error {
	fmt.Println(buildinfo.Version)
	if buildinfo.ReleaseNotes != "" {
		fmt.Printf("\nRelease Notes:\n%s\n", buildinfo.ReleaseNotes)
	}
	return nil
}

func (v *VersionListCommand) Run() error {
	versions, err := updater.FetchVersions()
	if err != nil {
		return fmt.Errorf("fetch versions: %w", err)
	}

	fmt.Println("Available Server Versions:")
	for _, ver := range versions.Server {
		fmt.Printf("- %s (%s)\n", ver.Version, ver.ReleaseDate)
	}

	fmt.Println("\nAvailable Agent Versions:")
	for _, ver := range versions.Agent {
		fmt.Printf("- %s (%s)\n", ver.Version, ver.ReleaseDate)
	}

	return nil
}

func (v *VersionRevertCommand) Run() error {
	manifest, err := updater.GetVersionManifestForRevert(v.Version, "")
	if err != nil {
		return err
	}

	fmt.Printf("Reverting to version %s...\n", v.Version)
	if err := updater.ApplyUpdate(manifest, false, buildinfo.Version); err != nil {
		return fmt.Errorf("apply update: %w", err)
	}

	_ = activity.RecordUserActivity(0, "Server Revert Applied (CLI)", "Reverted to version "+v.Version, "", "")

	if err := rpc.Call("Server.RestartServer", nil, nil); err == nil {
		fmt.Println("Chronix revert complete. Restarting Chronix daemon...")
	} else {
		fmt.Println("Chronix revert complete.")
	}

	return nil
}

func (b *BuildInfoCommand) Run() error {
	if bi, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("\nApp Name: %s\nGo Version: %s\nApp Version: %s\nPath: %s\nVersion: %s\n", consts.APPNAME, bi.GoVersion, buildinfo.Version, bi.Path, bi.Main.Version)
		for _, s := range bi.Settings {
			if strings.HasPrefix(s.Key, "-") {
				continue
			}
			fmt.Printf("%s: %s\n", s.Key, s.Value)
		}
		fmt.Printf("\n\n")
	} else {
		fmt.Println("no build information available")
	}
	return nil
}
