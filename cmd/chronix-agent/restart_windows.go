//go:build windows

package main

import (
	"os"
	"strings"
)

var PreRestart func()

func restart(exe string) error {
	if PreRestart != nil {
		PreRestart()
	}

	if agentServiceRunning && agentService != nil {
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
	if len(args) > 1 && args[1] == "register" {
		args = []string{args[0]}
	}

	attr := &os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	_, err := os.StartProcess(exe, args, attr)
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
