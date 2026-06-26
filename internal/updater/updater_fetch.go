package updater

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func CheckForUpdates(currentServerVersion string) (*VersionManifest, bool, error) {
	slog.Debug("Checking for updates", slog.String("url", ManifestURL))
	LastCheckTime = time.Now()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(ManifestURL)
	if err != nil {
		return nil, false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("failed to fetch manifest: %s", resp.Status)
	}

	var manifest VersionManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, false, err
	}

	AvailableVersion = &manifest
	return &manifest, IsNewerVersion(currentServerVersion, manifest.Server.Version), nil
}

func FetchVersions() (*VersionsManifest, error) {
	if CachedVersions != nil && time.Since(LastVersionsFetchTime) < 5*time.Minute {
		return CachedVersions, nil
	}

	slog.Debug("Fetching all available versions", slog.String("url", VersionsURL))
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(VersionsURL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch versions manifest: %s", resp.Status)
	}

	var manifest VersionsManifest
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	CachedVersions = &manifest
	LastVersionsFetchTime = time.Now()
	return &manifest, nil
}

func GetVersionManifestForRevert(serverVersion string, agentVersion string) (*VersionManifest, error) {
	versions, err := FetchVersions()
	if err != nil {
		return nil, err
	}

	manifest := &VersionManifest{}
	if serverVersion != "" {
		found := false
		for _, v := range versions.Server {
			if v.Version == serverVersion {
				manifest.Server = v
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("server version %s not found", serverVersion)
		}
	}

	if agentVersion != "" {
		found := false
		for _, v := range versions.Agent {
			if v.Version == agentVersion {
				manifest.Agent = v
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("agent version %s not found", agentVersion)
		}
	}

	return manifest, nil
}
