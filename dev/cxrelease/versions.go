package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func updateVersionsJSON(releaseDir, buildDir, privHex string) error {
	versionsPath := filepath.Join(releaseDir, "versions.json")
	var versions VersionsManifest

	if _, err := os.Stat(versionsPath); err == nil {
		vdata, err := os.ReadFile(versionsPath)
		if err == nil && len(vdata) > 0 {
			_ = json.Unmarshal(vdata, &versions)
		}
	}

	targets := []Target{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	entries, err := os.ReadDir(buildDir)
	if err != nil {
		return err
	}

	type binKey struct {
		relType string
		version string
	}
	presentBins := make(map[binKey]map[string]string)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		relType, ver, platform := parseBinaryName(entry.Name())
		if relType == "" {
			continue
		}
		key := binKey{relType, ver}
		if presentBins[key] == nil {
			presentBins[key] = make(map[string]string)
		}
		presentBins[key][platform] = filepath.Join(buildDir, entry.Name())
	}

	prune := func(list []VersionInfo, relType string) []VersionInfo {
		var newList []VersionInfo
		for _, info := range list {
			key := binKey{relType, info.Version}
			allPresent := true
			for _, t := range targets {
				platform := fmt.Sprintf("%s-%s", t.OS, t.Arch)
				if _, ok := presentBins[key][platform]; !ok {
					allPresent = false
					break
				}
			}
			if allPresent {
				newList = append(newList, info)
			} else {
				fmt.Printf("Pruning version %s (%s) from versions.json as some binaries are missing.\n", info.Version, relType)
			}
		}
		return newList
	}

	versions.Server = prune(versions.Server, "chronix")
	versions.Agent = prune(versions.Agent, "chronix-agent")
	versions.Tester = prune(versions.Tester, "chronix-tester")

	localPlatform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
	for key, bins := range presentBins {
		found := false
		var list *[]VersionInfo
		switch key.relType {
		case "chronix":
			list = &versions.Server
		case "chronix-agent":
			list = &versions.Agent
		case "chronix-tester":
			list = &versions.Tester
		default:
			continue
		}

		for _, v := range *list {
			if v.Version == key.version {
				found = true
				break
			}
		}
		if found {
			continue
		}

		allPresent := true
		for _, t := range targets {
			platform := fmt.Sprintf("%s-%s", t.OS, t.Arch)
			if _, ok := bins[platform]; !ok {
				allPresent = false
				break
			}
		}
		if !allPresent {
			continue
		}

		localPath, ok := bins[localPlatform]
		if !ok {
			fmt.Printf("Skipping supplementation for version %s (%s): no local binary for %s\n", key.version, key.relType, localPlatform)
			continue
		}

		fmt.Printf("Supplementing versions.json with version %s (%s)...\n", key.version, key.relType)
		ver, notes, err := getVersionInfoFromBinary(localPath)
		if err != nil {
			fmt.Printf("  Failed to get version info from %s: %v\n", localPath, err)
			continue
		}
		if ver != key.version {
			fmt.Printf("  Warning: version mismatch for %s. Found %s, expected %s\n", localPath, ver, key.version)
		}

		stat, _ := os.Stat(localPath)
		info := VersionInfo{
			Version:      key.version,
			ReleaseDate:  stat.ModTime().Format(time.RFC3339),
			ReleaseNotes: notes,
			Binaries:     make(map[string]BinaryInfo),
		}
		for platform, path := range bins {
			bInfo, err := getBinaryMetadata(path, privHex)
			if err != nil {
				fmt.Printf("  Failed to get metadata for %s: %v\n", path, err)
				continue
			}
			info.Binaries[platform] = bInfo
		}
		*list = append(*list, info)
	}

	vdata, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal versions: %w", err)
	}
	return os.WriteFile(versionsPath, vdata, 0644)
}

func verifyDarwinArmBinaries(buildDir string) {
	files, _ := filepath.Glob(filepath.Join(buildDir, "*-darwin-arm64"))
	for _, f := range files {
		cmd := exec.Command(f, "version")
		_, err := cmd.Output()
		if err != nil {
			fmt.Printf("Verification of %s FAILED: %v\n", filepath.Base(f), err)
		}
	}
}

func parseBinaryName(name string) (releaseType, version, platform string) {
	name = strings.TrimSuffix(name, ".exe")
	if strings.HasPrefix(name, "chronix-agent-") {
		releaseType = "chronix-agent"
		name = strings.TrimPrefix(name, "chronix-agent-")
	} else if strings.HasPrefix(name, "chronix-tester-") {
		releaseType = "chronix-tester"
		name = strings.TrimPrefix(name, "chronix-tester-")
	} else if strings.HasPrefix(name, "chronix-") {
		releaseType = "chronix"
		name = strings.TrimPrefix(name, "chronix-")
	} else if strings.HasPrefix(name, "agent-") {
		releaseType = "chronix-agent"
		name = strings.TrimPrefix(name, "agent-")
	} else if strings.HasPrefix(name, "tester-") {
		releaseType = "chronix-tester"
		name = strings.TrimPrefix(name, "tester-")
	} else {
		return "", "", ""
	}

	parts := strings.Split(name, "-")
	if len(parts) < 3 {
		return "", "", ""
	}

	arch := parts[len(parts)-1]
	osName := parts[len(parts)-2]
	platform = osName + "-" + arch
	version = strings.Join(parts[:len(parts)-2], "-")
	return
}

func getVersionInfoFromBinary(binaryPath string) (version, notes string, err error) {
	cmd := exec.Command(binaryPath, "version")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", "", err
	}

	lines := strings.Split(out.String(), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("empty output from version command")
	}

	notesStr := ""
	foundNotes := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "time=") && strings.Contains(trimmed, "level=") {
			continue
		}

		if version == "" {
			version = trimmed
			continue
		}
		if foundNotes {
			notesStr += line + "\n"
		} else if strings.HasPrefix(trimmed, "Release Notes:") {
			foundNotes = true
		}
	}

	return version, strings.TrimSpace(notesStr), nil
}
