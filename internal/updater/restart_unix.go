//go:build !windows

package updater

import (
	"chronix/internal/svc"
	"os"
	"runtime"
	"strings"
	"syscall"
)

func Restart(exe string) error {
	if PreRestart != nil {
		PreRestart()
	}

	if svc.ServiceRunning && svc.Svc != nil && runtime.GOOS != "darwin" {
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

	// Use syscall.Exec to replace the current process. This is important
	// for shell-initiated runs so they stay in the foreground.
	return syscall.Exec(exe, os.Args, os.Environ())
}
