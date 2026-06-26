package updater

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VersionManifest struct {
	Server VersionInfo `json:"server"`
	Agent  VersionInfo `json:"chronix-agent"`
}

type VersionsManifest struct {
	Server []VersionInfo `json:"server"`
	Agent  []VersionInfo `json:"chronix-agent"`
}

type VersionInfo struct {
	Version      string                `json:"version"`
	ReleaseDate  string                `json:"release_date"`
	ReleaseNotes string                `json:"release_notes"`
	Binaries     map[string]BinaryInfo `json:"binaries"`
}

type BinaryInfo struct {
	URL       string `json:"url"`
	SHA256    string `json:"sha256"`
	Signature string `json:"signature"`
}

var (
	Enabled       = true
	Mode          = "notify" // notify, automatic
	WindowStart   = ""
	CheckInterval = 6 * time.Hour
	ManifestURL   = "https://dist.chronixhq.com/latest.json"
	VersionsURL   = "https://dist.chronixhq.com/versions.json"
	PublicKey     = "01b4ea575c9ecae3dbea0e87f46c9a6cab91fe521f8c560156d0690f8817e9fa" // Ed25519 public key in hex. If set, binary signatures will be strictly verified.
	DataDir       = ""

	AgentEnabled     = true
	AgentMode        = "notify" // notify, automatic
	AgentWindowStart = ""
)

var (
	LastCheckTime    time.Time
	AvailableVersion *VersionManifest

	LastVersionsFetchTime time.Time
	CachedVersions        *VersionsManifest

	// TriggerAgentUpdate is a hook set by the server to notify agents about updates.
	TriggerAgentUpdate func(version string)
	// RecordActivityHook is a hook set by the server to record update activity.
	RecordActivityHook func(action string, details string)
	// PreRestart is a hook called before the process restarts.
	PreRestart func()
)

func getTargetExePath(currentPath, currentVersion, newVersion string) string {
	dir := filepath.Dir(currentPath)
	base := filepath.Base(currentPath)

	if currentVersion != "" && currentVersion != "dev" && strings.Contains(base, currentVersion) {
		newBase := strings.Replace(base, currentVersion, newVersion, 1)
		return filepath.Join(dir, newBase)
	}

	return currentPath
}

func CleanupOldExecutable() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exePath)
	base := filepath.Base(exePath)

	// Standard check for [current].old
	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		slog.Info("Removing old executable backup", slog.String("path", oldPath))
		if err := os.Remove(oldPath); err != nil {
			slog.Warn("Failed to remove old executable backup", slog.String("path", oldPath), slog.String("error", err.Error()))
		}
	}

	// Also look for any other .old files in the same directory that might be from a versioned update.
	// We only clean up files that start with our prefix (e.g. "chronix-" or "chronix-agent-")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	prefix := base
	if strings.HasPrefix(base, "chronix-agent-") {
		prefix = "chronix-agent-"
	} else if idx := strings.Index(base, "-"); idx > 0 {
		prefix = base[:idx+1]
	}

	for _, f := range entries {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if strings.HasSuffix(name, ".old") && strings.HasPrefix(name, prefix) {
			fullPath := filepath.Join(dir, name)
			if fullPath == oldPath {
				continue // already tried
			}
			slog.Info("Removing old versioned executable backup", slog.String("path", fullPath))
			if err := os.Remove(fullPath); err != nil {
				slog.Warn("Failed to remove old versioned executable backup", slog.String("path", fullPath), slog.String("error", err.Error()))
			}
		}
	}
}
