// Package main provides a tool for releasing and distributing Chronix and Agent binaries.
package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type KeyPair struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}

type VersionManifest struct {
	Server VersionInfo `json:"server"`
	Agent  VersionInfo `json:"chronix-agent"`
	Tester VersionInfo `json:"chronix-tester"`
}

type VersionsManifest struct {
	Server []VersionInfo `json:"server"`
	Agent  []VersionInfo `json:"chronix-agent"`
	Tester []VersionInfo `json:"chronix-tester"`
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

type Target struct {
	OS   string
	Arch string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	if command == "deploy" {
		version := ""
		if len(os.Args) >= 3 {
			version = os.Args[2]
		}
		deploy(version)
		return
	}

	if len(os.Args) < 4 {
		printUsage()
		os.Exit(1)
	}

	releaseType := command
	version := os.Args[2]
	releaseNotes := os.Args[3]

	// Determine directories
	_, filename, _, _ := runtime.Caller(0)
	releaseDir := filepath.Dir(filename)
	projectRoot := filepath.Dir(filepath.Dir(releaseDir))
	buildDir := filepath.Join(releaseDir, "builds")

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		fmt.Printf("Failed to create builds directory: %v\n", err)
		os.Exit(1)
	}

	if version == "maint" || version == "minor" || version == "major" {
		manifestPath := filepath.Join(releaseDir, "latest.json")
		data, err := os.ReadFile(manifestPath)
		if err != nil {
			if os.IsNotExist(err) || len(data) == 0 {
				data = []byte("{}")
			} else {
				fmt.Printf("Failed to read latest.json for version bumping: %v\n", err)
				os.Exit(1)
			}
		}
		var manifest VersionManifest
		if err := json.Unmarshal(data, &manifest); err != nil && len(data) > 0 {
			fmt.Printf("Failed to unmarshal latest.json: %v\n", err)
			os.Exit(1)
		}

		currentVersion := ""
		switch releaseType {
		case "chronix":
			currentVersion = manifest.Server.Version
		case "chronix-agent", "agent":
			currentVersion = manifest.Agent.Version
		case "chronix-tester", "tester":
			currentVersion = manifest.Tester.Version
		}

		if currentVersion == "" {
			currentVersion = "0.0.0"
		}

		newVersion, err := bumpVersion(currentVersion, version)
		if err != nil {
			fmt.Printf("Failed to bump version: %v\n", err)
			os.Exit(1)
		}
		version = newVersion
		fmt.Printf("Auto-incremented %s version to %s\n", releaseType, version)
	}

	keyPath := filepath.Join(releaseDir, "chronix.keys")
	if len(os.Args) > 4 {
		keyPath = os.Args[4]
	}

	if releaseType != "chronix" && releaseType != "agent" && releaseType != "chronix-agent" && releaseType != "tester" && releaseType != "chronix-tester" {
		fmt.Printf("Invalid release type: %s. Must be 'chronix', 'chronix-agent' or 'chronix-tester'\n", releaseType)
		os.Exit(1)
	}

	if releaseType == "agent" {
		releaseType = "chronix-agent"
	}
	if releaseType == "tester" {
		releaseType = "chronix-tester"
	}

	// 1. Key Management
	priv, pub, err := loadOrGenKeys(keyPath, len(os.Args) > 4)
	if err != nil {
		fmt.Printf("Key error: %v\n", err)
		os.Exit(1)
	}

	pubHex := hex.EncodeToString(pub)
	privHex := hex.EncodeToString(priv)

	// 2. Release for all platforms
	targets := []Target{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "amd64"},
		{"darwin", "arm64"},
		{"windows", "amd64"},
	}

	manifestPath := filepath.Join(releaseDir, "latest.json")

	// 2. Build React App (for chronix only)
	if releaseType == "chronix" {
		reactDir := filepath.Join(projectRoot, "internal", "chronix_react_app")

		// Update package.json version
		fmt.Printf("Updating React package.json to version %s...\n", version)
		packageJSONPath := filepath.Join(reactDir, "package.json")
		if err := updatePackageVersion(packageJSONPath, version); err != nil {
			fmt.Printf("Failed to update package.json: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Building React application...")
		cmd := exec.Command("npm", "run", "build")
		cmd.Dir = reactDir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("React build failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("React application build successful.")

		// 2.5. Rebuild schema.db from dev/schema.sql
		fmt.Println("Rebuilding schema.db from dev/schema.sql...")
		schemaSQLPath := filepath.Join(projectRoot, "dev", "schema.sql")
		schemaDBPath := filepath.Join(projectRoot, "internal", "db", "assets", "schema.db")

		// Remove existing schema.db to start fresh
		_ = os.Remove(schemaDBPath)

		schemaCmd := exec.Command("sqlite3", schemaDBPath)
		schemaCmd.Stdin = strings.NewReader(fmt.Sprintf(".read %s", schemaSQLPath))
		if err := schemaCmd.Run(); err != nil {
			fmt.Printf("Failed to rebuild schema.db: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("schema.db rebuilt successfully.")
	}

	for _, target := range targets {
		platform := fmt.Sprintf("%s-%s", target.OS, target.Arch)
		binaryName := fmt.Sprintf("%s-%s-%s", releaseType, version, platform)
		if target.OS == "windows" {
			binaryName += ".exe"
		}

		binaryPath := filepath.Join(buildDir, binaryName)

		fmt.Printf("Building %s version %s for %s...\n", releaseType, version, platform)

		var ldflags string
		var pkgRelPath string
		var workDir string
		switch releaseType {
		case "chronix":
			ldflags = fmt.Sprintf("-X chronix/pkg/buildinfo.Version=%s -X 'chronix/pkg/buildinfo.ReleaseNotes=%s' -X chronix/internal/updater.PublicKey=%s", version, releaseNotes, pubHex)
			pkgRelPath = "./cmd"
			workDir = projectRoot
		case "chronix-agent":
			ldflags = fmt.Sprintf("-X main.Version=%s -X 'main.ReleaseNotes=%s' -X main.PublicKey=%s", version, releaseNotes, pubHex)
			pkgRelPath = "."
			workDir = filepath.Join(projectRoot, "cmd", "chronix-agent")
		case "chronix-tester":
			ldflags = fmt.Sprintf("-X main.Version=%s -X 'main.ReleaseNotes=%s'", version, releaseNotes)
			pkgRelPath = "."
			workDir = filepath.Join(projectRoot, "dev", "chronix_tester")
		}

		cmd := exec.Command("go", "build", "-ldflags", ldflags, "-o", binaryPath, pkgRelPath)
		cmd.Dir = workDir
		cmd.Env = append(os.Environ(), "GOOS="+target.OS, "GOARCH="+target.Arch, "CGO_ENABLED=0")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Build failed for %s: %v\n", platform, err)
			os.Exit(1)
		}

		fmt.Printf("Binary created: %s\n", binaryName)

		// 3. Manifest Update
		if err := updateManifest(manifestPath, releaseType, version, releaseNotes, binaryPath, platform, privHex); err != nil {
			fmt.Printf("Manifest update failed for %s: %v\n", platform, err)
			os.Exit(1)
		}
	}

	// 4. Update versions.json (prune and supplement)
	if err := updateVersionsJSON(releaseDir, buildDir, privHex); err != nil {
		fmt.Printf("Versions manifest update failed: %v\n", err)
		os.Exit(1)
	}

	// 5. Verify darwin-arm64 binaries
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		fmt.Println("\nRunning verification for darwin-arm64 binaries...")
		verifyDarwinArmBinaries(buildDir)
	}

	fmt.Println("Release for all platforms and manifest update completed successfully.")
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  Build and sign:")
	fmt.Println("    go run dev/cxrelease/main.go <chronix|chronix-agent|chronix-tester> <version|maint|minor|major> <release notes> [key_path]")
	fmt.Println("  Deploy to indomitus:")
	fmt.Println("    go run dev/cxrelease/main.go deploy [version]")
}
