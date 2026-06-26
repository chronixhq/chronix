package main

import (
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultConfigPath_XDG(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Skipping XDG test on non-linux")
	}

	oldXDG := os.Getenv("XDG_DATA_HOME")
	defer func() { _ = os.Setenv("XDG_DATA_HOME", oldXDG) }()

	// Test non-existent XDG_DATA_HOME
	_ = os.Setenv("XDG_DATA_HOME", "/non/existent/path")
	p := defaultConfigPath()
	if strings.Contains(p, "/non/existent/path") {
		t.Errorf("defaultConfigPath should have ignored non-existent XDG_DATA_HOME: %s", p)
	}

	// Test accessible XDG_DATA_HOME
	tmpDir, err := os.MkdirTemp("", "xdg_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	_ = os.Setenv("XDG_DATA_HOME", tmpDir)
	p = defaultConfigPath()
	if !strings.Contains(p, tmpDir) {
		t.Errorf("defaultConfigPath should have used XDG_DATA_HOME: %s", p)
	}

	// Test inaccessible XDG_DATA_HOME (if possible)
	// On Linux, we can check if it belongs to us.
	// But mocking ownership is hard.
}
