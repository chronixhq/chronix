// Package buildinfo exposes application build metadata injected at build time.
package buildinfo

var (
	// Version is the application version injected at build time.
	Version = "dev"
	// ReleaseNotes carries optional release notes injected at build time.
	ReleaseNotes = ""
)
