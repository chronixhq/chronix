//go:build windows

package updater

import (
	"chronix/internal/svc"
	"os"
	"strings"
)

func Restart(exe string) error {
	if PreRestart != nil {
		PreRestart()
	}

	if svc.ServiceRunning && svc.Svc != nil {
		return svc.Svc.Restart()
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

	attr := &os.ProcAttr{
		Dir:   ".",
		Env:   os.Environ(),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	_, err := os.StartProcess(exe, os.Args, attr)
	if err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
