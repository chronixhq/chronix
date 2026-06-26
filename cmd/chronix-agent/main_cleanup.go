package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func cleanupOldExecutable() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exePath)
	base := filepath.Base(exePath)

	oldPath := exePath + ".old"
	if _, err := os.Stat(oldPath); err == nil {
		slog.Info("Removing old executable backup", "path", oldPath)
		if err := os.Remove(oldPath); err != nil {
			slog.Warn("Failed to remove old executable backup", "path", oldPath, "error", err)
		}
	}

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
				continue
			}
			slog.Info("Removing old versioned executable backup", "path", fullPath)
			if err := os.Remove(fullPath); err != nil {
				slog.Warn("Failed to remove old versioned executable backup", "path", fullPath, "error", err)
			}
		}
	}
}
