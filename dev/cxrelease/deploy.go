package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func deploy(version string) {
	_, filename, _, _ := runtime.Caller(0)
	releaseDir := filepath.Dir(filename)
	buildDir := filepath.Join(releaseDir, "builds")

	if version != "" {
		fmt.Printf("Deploying version %s to indomitus...\n", version)
	} else {
		fmt.Println("Deploying all versions to indomitus...")
	}

	var patterns []string
	if version != "" {
		patterns = []string{
			filepath.Join(buildDir, fmt.Sprintf("chronix-%s-*", version)),
			filepath.Join(buildDir, fmt.Sprintf("agent-%s-*", version)),
			filepath.Join(buildDir, fmt.Sprintf("chronix-agent-%s-*", version)),
			filepath.Join(buildDir, fmt.Sprintf("chronix-tester-%s-*", version)),
		}
	} else {
		patterns = []string{
			filepath.Join(buildDir, "chronix-*"),
			filepath.Join(buildDir, "agent-*"),
			filepath.Join(buildDir, "chronix-agent-*"),
			filepath.Join(buildDir, "chronix-tester-*"),
		}
	}

	var allFiles []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			fmt.Printf("Glob error for %s: %v\n", pattern, err)
			continue
		}
		allFiles = append(allFiles, matches...)
	}

	if len(allFiles) == 0 {
		if version != "" {
			fmt.Printf("No binaries found for version %s in %s\n", version, buildDir)
		} else {
			fmt.Printf("No binaries found in %s\n", buildDir)
		}
		os.Exit(1)
	}

	remoteBaseDir := "/web/dist.chronixhq.com"
	remoteDownloadsDir := fmt.Sprintf("%s/downloads", remoteBaseDir)

	fmt.Printf("Ensuring remote directory exists: %s\n", remoteDownloadsDir)
	mkdirCmd := exec.Command("ssh", "root@indomitus", fmt.Sprintf("mkdir -p %s", remoteDownloadsDir))
	mkdirCmd.Stdout = os.Stdout
	mkdirCmd.Stderr = os.Stderr
	if err := mkdirCmd.Run(); err != nil {
		fmt.Printf("Failed to create remote directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Syncing binaries...")
	args := []string{"-avz"}
	args = append(args, allFiles...)
	args = append(args, fmt.Sprintf("root@indomitus:%s/", remoteDownloadsDir))
	rsyncBinCmd := exec.Command("rsync", args...)
	rsyncBinCmd.Stdout = os.Stdout
	rsyncBinCmd.Stderr = os.Stderr
	if err := rsyncBinCmd.Run(); err != nil {
		fmt.Printf("Rsync binaries failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Syncing latest.json...")
	manifestPath := filepath.Join(releaseDir, "latest.json")
	rsyncManifestCmd := exec.Command("rsync", "-avz", manifestPath, fmt.Sprintf("root@indomitus:%s/", remoteBaseDir))
	rsyncManifestCmd.Stdout = os.Stdout
	rsyncManifestCmd.Stderr = os.Stderr
	if err := rsyncManifestCmd.Run(); err != nil {
		fmt.Printf("Rsync latest.json failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Syncing versions.json...")
	versionsPath := filepath.Join(releaseDir, "versions.json")
	if _, err := os.Stat(versionsPath); err == nil {
		rsyncVersionsCmd := exec.Command("rsync", "-avz", versionsPath, fmt.Sprintf("root@indomitus:%s/", remoteBaseDir))
		rsyncVersionsCmd.Stdout = os.Stdout
		rsyncVersionsCmd.Stderr = os.Stderr
		if err := rsyncVersionsCmd.Run(); err != nil {
			fmt.Printf("Rsync versions.json failed: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("Syncing install.sh...")
	installScriptPath := filepath.Join(releaseDir, "install.sh")
	if _, err := os.Stat(installScriptPath); err == nil {
		rsyncInstallCmd := exec.Command("rsync", "-avz", installScriptPath, fmt.Sprintf("root@indomitus:%s/", remoteBaseDir))
		rsyncInstallCmd.Stdout = os.Stdout
		rsyncInstallCmd.Stderr = os.Stderr
		if err := rsyncInstallCmd.Run(); err != nil {
			fmt.Printf("Rsync install.sh failed: %v\n", err)
		}
	}

	latestVersion := version
	if latestVersion == "" {
		manifestPath := filepath.Join(releaseDir, "latest.json")
		if data, err := os.ReadFile(manifestPath); err == nil {
			var manifest VersionManifest
			if err := json.Unmarshal(data, &manifest); err == nil {
				latestVersion = manifest.Server.Version
			}
		}
	}

	if latestVersion != "" {
		fmt.Printf("Updating 'latest' pointers for version %s...\n", latestVersion)
		targets := []Target{
			{"linux", "amd64"},
			{"linux", "arm64"},
			{"darwin", "amd64"},
			{"darwin", "arm64"},
			{"windows", "amd64"},
		}
		for _, target := range targets {
			platform := fmt.Sprintf("%s-%s", target.OS, target.Arch)
			remotePlatformDir := fmt.Sprintf("%s/latest/%s", remoteBaseDir, platform)

			binaryName := fmt.Sprintf("chronix-%s-%s", latestVersion, platform)
			if target.OS == "windows" {
				binaryName += ".exe"
			}

			destBinary := "chronix"
			if target.OS == "windows" {
				destBinary = "chronix.exe"
			}

			checkCmd := fmt.Sprintf("[ -f %s/%s ]", remoteDownloadsDir, binaryName)
			if err := exec.Command("ssh", "root@indomitus", checkCmd).Run(); err == nil {
				fmt.Printf("Updating %s/latest/%s/%s -> %s\n", remoteBaseDir, platform, destBinary, binaryName)
				updateCmd := fmt.Sprintf("mkdir -p %s && cp %s/%s %s/%s", remotePlatformDir, remoteDownloadsDir, binaryName, remotePlatformDir, destBinary)
				if err := exec.Command("ssh", "root@indomitus", updateCmd).Run(); err != nil {
					fmt.Printf("Failed to update latest for %s: %v\n", platform, err)
				}
			}
		}
	}

	if version != "" {
		fmt.Printf("Deployment of version %s to indomitus completed successfully.\n", version)
	} else {
		fmt.Println("Deployment of all versions to indomitus completed successfully.")
	}
}
