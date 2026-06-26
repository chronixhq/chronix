//go:build windows

package fileutil

import "os"

func IsOwnedByUser(info os.FileInfo) bool {
	return true
}
