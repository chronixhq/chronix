package main

import (
	updatepkg "chronix-agent/agentupdate"
	"log/slog"
	"os"
)

func agentDownloadAndApplyUpdate(cfg *agentConfig, version string, sha256Expected string, signatureExpected string) error {
	restartOrExit := func(path string) error {
		if err := restart(path); err != nil {
			slog.Error("Failed to restart agent after update", "error", err)
			os.Exit(1)
		}
		return nil
	}

	return updatepkg.ApplyUpdate(cfg, version, sha256Expected, signatureExpected, Version, PublicKey, restartOrExit)
}
