package rpc

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDefaultSocketPath_Precedence(t *testing.T) {
	oldRPC := os.Getenv("RPC_SOCKET_PATH")
	oldXDG := os.Getenv("XDG_RUNTIME_DIR")
	defer func() {
		_ = os.Setenv("RPC_SOCKET_PATH", oldRPC)
		_ = os.Setenv("XDG_RUNTIME_DIR", oldXDG)
	}()

	_ = os.Unsetenv("RPC_SOCKET_PATH")
	_ = os.Unsetenv("XDG_RUNTIME_DIR")
	p := DefaultSocketPath()
	if filepath.Base(p) != "chronix-rpc.sock" {
		t.Fatalf("unexpected default base: %s", p)
	}

	_ = os.Setenv("XDG_RUNTIME_DIR", "/tmp/xdg")
	// Note: We need to ensure /tmp/xdg exists and is a directory for DefaultSocketPath to pick it up now
	_ = os.MkdirAll("/tmp/xdg", 0755)
	defer func() { _ = os.RemoveAll("/tmp/xdg") }()
	p = DefaultSocketPath()
	if p != filepath.Join("/tmp/xdg", "chronix", "chronix-rpc.sock") {
		t.Fatalf("unexpected XDG path: %s", p)
	}

	// Test non-existent XDG_RUNTIME_DIR
	_ = os.Setenv("XDG_RUNTIME_DIR", "/non/existent/path")
	p = DefaultSocketPath()
	if strings.Contains(p, "/non/existent/path") {
		t.Fatalf("DefaultSocketPath should have ignored non-existent XDG_RUNTIME_DIR: %s", p)
	}
	if filepath.Base(p) != "chronix-rpc.sock" {
		t.Fatalf("unexpected fallback base: %s", p)
	}

	_ = os.Setenv("RPC_SOCKET_PATH", "/custom/path.sock")
	p = DefaultSocketPath()
	if p != "/custom/path.sock" {
		t.Fatalf("RPC_SOCKET_PATH should win: %s", p)
	}

	if runtime.GOOS == "windows" {
		// Just ensure it returns a path without panic
		_ = DefaultSocketPath()
	}
}
