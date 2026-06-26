//go:build !windows

// Package fileutil provides file-related utility functions.
package fileutil

import (
	"os"
	"syscall"
)

// IsOwnedByUser returns true if the file is owned by the current user.
// On Unix-like systems, it checks the UID. On non-syscall systems, it returns true.
func IsOwnedByUser(info os.FileInfo) bool {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return stat.Uid == uint32(os.Getuid())
	}
	return true
}
