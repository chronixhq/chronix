package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type AppPaths struct {
	DataDir   string
	ResultsDB string
	TargetDB  string
}

type ShellCommandSpec struct {
	Command    string
	WorkingDir *string
}

func resolvePaths(explicit string) (AppPaths, error) {
	dataDir, err := resolveDataDir(explicit)
	if err != nil {
		return AppPaths{}, err
	}
	return AppPaths{
		DataDir:   dataDir,
		ResultsDB: filepath.Join(dataDir, "results.db"),
		TargetDB:  filepath.Join(dataDir, "target.db"),
	}, nil
}

func resolveDataDir(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return filepath.Abs(explicit)
	}

	if envDir := strings.TrimSpace(os.Getenv("CHRONIX_TESTER_DATA_DIR")); envDir != "" {
		return filepath.Abs(envDir)
	}

	if dir := sourceTesterDir(); dir != "" {
		return dir, nil
	}

	if exe, err := os.Executable(); err == nil {
		if dir := filepath.Dir(exe); dirExists(dir) {
			return filepath.Abs(dir)
		}
	}

	return os.Getwd()
}

func sourceTesterDir() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	dir := filepath.Dir(file)
	if fileExists(filepath.Join(dir, "go.mod")) {
		return dir
	}
	return ""
}

func sourceRepoRoot() string {
	testerDir := sourceTesterDir()
	if testerDir == "" {
		return ""
	}
	devDir := filepath.Dir(testerDir)
	if filepath.Base(devDir) != "dev" {
		return ""
	}
	root := filepath.Dir(devDir)
	if fileExists(filepath.Join(root, "go.mod")) {
		return root
	}
	return ""
}

func resolveTokenCommand(paths AppPaths, override string, overrideWorkingDir string) ShellCommandSpec {
	if strings.TrimSpace(override) != "" {
		spec := ShellCommandSpec{Command: override}
		if strings.TrimSpace(overrideWorkingDir) != "" {
			wd := overrideWorkingDir
			spec.WorkingDir = &wd
		}
		return spec
	}

	candidates := []string{
		filepath.Join(paths.DataDir, "chronix-tester"),
		filepath.Join(paths.DataDir, "chronix_tester"),
	}
	if dir := sourceTesterDir(); dir != "" {
		candidates = append(candidates,
			filepath.Join(dir, "chronix-tester"),
			filepath.Join(dir, "chronix_tester"),
		)
	}

	for _, candidate := range candidates {
		if fileExists(candidate) {
			return ShellCommandSpec{
				Command: fmt.Sprintf("%s generate-token --data-dir %s", quoteShellArg(candidate), quoteShellArg(paths.DataDir)),
			}
		}
	}

	if exe, err := os.Executable(); err == nil && stableExecutable(exe) {
		return ShellCommandSpec{
			Command: fmt.Sprintf("%s generate-token --data-dir %s", quoteShellArg(exe), quoteShellArg(paths.DataDir)),
		}
	}

	if root := sourceRepoRoot(); root != "" {
		wd := root
		return ShellCommandSpec{
			Command:    fmt.Sprintf("go run ./dev/chronix_tester generate-token --data-dir %s", quoteShellArg(paths.DataDir)),
			WorkingDir: &wd,
		}
	}

	return ShellCommandSpec{
		Command: fmt.Sprintf("chronix-tester generate-token --data-dir %s", quoteShellArg(paths.DataDir)),
	}
}

func stableExecutable(path string) bool {
	if !fileExists(path) {
		return false
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	tempDir := filepath.Clean(os.TempDir())
	rel, err := filepath.Rel(tempDir, absPath)
	if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return false
	}
	return true
}

func quoteShellArg(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
