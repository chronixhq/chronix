package shellrun

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// getShell returns a reasonable shell path for the current OS.
func getShell() string {
	if runtime.GOOS == "windows" {
		// Minimal support: use PowerShell if available; however, most CI for this project is Linux/macOS.
		// We still return "powershell" so tests are skipped on Windows below.
		return "powershell"
	}
	return "/bin/sh"
}

func TestTruncateBytes(t *testing.T) {
	b := []byte("abcdefghijklmnopqrstuvwxyz")
	// No truncate
	seg, trunc, total := TruncateBytes(b, 0, "head")
	if trunc || total != len(b) || string(seg) != string(b) {
		t.Fatalf("expected no truncate; trunc=%v total=%d seg=%q", trunc, total, string(seg))
	}
	// Head
	seg, trunc, total = TruncateBytes(b, 5, "head")
	if !trunc || total != len(b) || string(seg) != "abcde" {
		t.Fatalf("head truncate mismatch: trunc=%v total=%d seg=%q", trunc, total, string(seg))
	}
	// Tail
	seg, trunc, total = TruncateBytes(b, 3, "tail")
	if !trunc || total != len(b) || string(seg) != "xyz" {
		t.Fatalf("tail truncate mismatch: trunc=%v total=%d seg=%q", trunc, total, string(seg))
	}
}

func TestRunLocal_CommandEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("RunLocal tests target POSIX shells in CI")
	}
	ctx := context.Background()
	msg := "hello-world"
	cmd := "echo " + msg
	code, stdout, stderr, err := RunLocal(ctx, getShell(), "command", &cmd, nil, nil, map[string]string{}, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr=%s)", err, string(stderr))
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr=%s)", code, string(stderr))
	}
	if got := strings.TrimSpace(string(stdout)); got != msg {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestRunLocal_ScriptExitAndStderr(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("RunLocal tests target POSIX shells in CI")
	}
	ctx := context.Background()
	s := "echo hi; echo oops 1>&2; exit 5"
	code, stdout, stderr, err := RunLocal(ctx, getShell(), "script", nil, &s, nil, map[string]string{}, false, nil, nil)
	if err == nil {
		// exec.Command returns error on non-zero exit
		t.Fatalf("expected error from non-zero exit, got nil")
	}
	if code != 5 {
		t.Fatalf("expected exit code 5, got %d", code)
	}
	if !strings.Contains(string(stdout), "hi") {
		t.Fatalf("stdout missing content: %q", string(stdout))
	}
	if !strings.Contains(string(stderr), "oops") {
		t.Fatalf("stderr missing content: %q", string(stderr))
	}
}

func TestRunLocal_WorkingDirAndEnv(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("RunLocal tests target POSIX shells in CI")
	}
	dir := t.TempDir()
	// Verify PWD equals workingDir and env variable is visible
	cmd := "printf '%s %s' \"$PWD\" \"$FOO\""
	env := map[string]string{"FOO": "bar"}
	code, stdout, stderr, err := RunLocal(context.Background(), getShell(), "command", &cmd, nil, &dir, env, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr=%s)", err, string(stderr))
	}
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d (stderr=%s)", code, string(stderr))
	}
	out := strings.TrimSpace(string(stdout))
	// stdout should be: "<dir> bar" — account for macOS /private prefix on TMP dirs
	wantPrefix := dir + " "
	if runtime.GOOS == "darwin" && strings.HasPrefix(out, "/private") && !strings.HasPrefix(dir, "/private") {
		wantPrefix = "/private" + wantPrefix
	}
	if !strings.HasPrefix(out, wantPrefix) {
		t.Fatalf("expected PWD prefix %q, got %q", wantPrefix, out)
	}
	if !strings.HasSuffix(out, " bar") {
		t.Fatalf("expected env suffix ' bar', got %q", out)
	}

	// Also ensure the working dir exists as created by RunLocal when missing
	missing := filepath.Join(dir, "subdir/child")
	cmdTouch := "mkdir -p subdir/child && test -d subdir/child"
	code, stdout, stderr, err = RunLocal(context.Background(), getShell(), "command", &cmdTouch, nil, &dir, map[string]string{}, false, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v (stderr=%s stdout=%s)", err, string(stderr), string(stdout))
	}
	if code != 0 {
		t.Fatalf("expected exit code 0 for mkdir/test, got %d (stderr=%s)", code, string(stderr))
	}
	if _, statErr := os.Stat(missing); statErr != nil {
		t.Fatalf("expected path to exist: %s, err=%v", missing, statErr)
	}
}

func TestRunLocal_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("RunLocal tests target POSIX shells in CI")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	cmd := "sleep 2"
	code, _, _, err := RunLocal(ctx, getShell(), "command", &cmd, nil, nil, map[string]string{}, false, nil, nil)
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
	if code != 124 {
		t.Fatalf("expected exit code 124 on timeout, got %d", code)
	}
}
