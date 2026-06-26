//go:build !windows

package main

import (
	"os"
	"runtime"
	"strings"
	"syscall"
)

var PreRestart func()

func restart(exe string) error {
	if PreRestart != nil {
		PreRestart()
	}

	if agentServiceRunning && agentService != nil && runtime.GOOS != "darwin" {
		return agentService.Restart()
	}

	if exe == "" {
		var err error
		exe, err = os.Executable()
		if err != nil {
			return err
		}

		// If we are running from a .old file, try to use the one without .old
		// as it's likely been replaced by a new binary.
		if strings.HasSuffix(exe, ".old") {
			primary := strings.TrimSuffix(exe, ".old")
			if _, err := os.Stat(primary); err == nil {
				exe = primary
			}
		}
	}

	args := os.Args
	// Strip "register" subcommand if present to avoid re-registration loops
	if len(args) > 1 && args[1] == "register" {
		args = []string{args[0]}
	}

	// Use syscall.Exec to replace the current process. This is important
	// for shell-initiated runs so they stay in the foreground.
	return syscall.Exec(exe, args, os.Environ())
}
