package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func updateManifest(manifestPath, releaseType, version, releaseNotes, binaryPath, platform, privHex string) error {
	var manifest VersionManifest

	if _, err := os.Stat(manifestPath); err == nil {
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			return fmt.Errorf("read manifest: %w", err)
		}
		if err := json.Unmarshal(data, &manifest); err != nil && len(data) > 0 {
			return fmt.Errorf("unmarshal manifest: %w", err)
		}
	} else {
		manifest.Server.Binaries = make(map[string]BinaryInfo)
		manifest.Agent.Binaries = make(map[string]BinaryInfo)
		manifest.Tester.Binaries = make(map[string]BinaryInfo)
	}

	info, err := getBinaryMetadata(binaryPath, privHex)
	if err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339)
	switch releaseType {
	case "chronix":
		manifest.Server.Version = version
		manifest.Server.ReleaseDate = now
		manifest.Server.ReleaseNotes = releaseNotes
		if manifest.Server.Binaries == nil {
			manifest.Server.Binaries = make(map[string]BinaryInfo)
		}
		manifest.Server.Binaries[platform] = info
	case "chronix-agent":
		manifest.Agent.Version = version
		manifest.Agent.ReleaseDate = now
		manifest.Agent.ReleaseNotes = releaseNotes
		if manifest.Agent.Binaries == nil {
			manifest.Agent.Binaries = make(map[string]BinaryInfo)
		}
		manifest.Agent.Binaries[platform] = info
	case "chronix-tester":
		manifest.Tester.Version = version
		manifest.Tester.ReleaseDate = now
		manifest.Tester.ReleaseNotes = releaseNotes
		if manifest.Tester.Binaries == nil {
			manifest.Tester.Binaries = make(map[string]BinaryInfo)
		}
		manifest.Tester.Binaries[platform] = info
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0644)
}

func updatePackageVersion(packagePath, version string) error {
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return fmt.Errorf("read package.json: %w", err)
	}

	var pkg map[string]any
	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("unmarshal package.json: %w", err)
	}
	pkg["version"] = version

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(pkg); err != nil {
		return fmt.Errorf("marshal package.json: %w", err)
	}

	if err := os.WriteFile(packagePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write package.json: %w", err)
	}
	return nil
}
