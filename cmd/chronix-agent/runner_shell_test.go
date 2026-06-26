package main

import (
	runpkg "chronix-agent/agentrun"
	"context"
	"runtime"
	"strings"
	"testing"
)

func TestAgentDefaultShellPath(t *testing.T) {
	got := runpkg.DefaultShellPath("windows")
	if !strings.HasSuffix(strings.ToLower(got), "powershell.exe") {
		t.Fatalf("windows default shell path: got %q, want it to end with %q", got, "powershell.exe")
	}
	if got := runpkg.DefaultShellPath("linux"); got != "/bin/sh" {
		t.Fatalf("linux default shell path: got %q, want %q", got, "/bin/sh")
	}
}

func TestAgentShellCommandInvocationArgs_WindowsPowerShell(t *testing.T) {
	shellPath := runpkg.DefaultShellPath("windows")
	args := runpkg.ShellCommandInvocationArgs("windows", shellPath, "echo ok", map[string]string{"FOO": "bar"})
	if len(args) < 4 {
		t.Fatalf("args too short: %#v", args)
	}
	if !strings.HasSuffix(strings.ToLower(args[0]), "powershell.exe") {
		t.Fatalf("shell: got %q, want it to end with %q", args[0], "powershell.exe")
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "-Command") {
		t.Fatalf("expected -Command in args: %#v", args)
	}
	if strings.Contains(joined, "-c") {
		t.Fatalf("did not expect POSIX -c in Windows args: %#v", args)
	}
}

func TestAgentShellCommandInvocationArgs_WindowsCmd(t *testing.T) {
	args := runpkg.ShellCommandInvocationArgs("windows", "cmd.exe", "echo ok", nil)
	if len(args) != 3 {
		t.Fatalf("unexpected args: %#v", args)
	}
	if args[1] != "/C" {
		t.Fatalf("expected /C for cmd.exe, got: %#v", args)
	}
}

func TestAgentShellCommandInvocationArgs_POSIX(t *testing.T) {
	args := runpkg.ShellCommandInvocationArgs("linux", "/bin/sh", "echo ok", map[string]string{"A": "1"})
	if len(args) != 3 {
		t.Fatalf("unexpected args: %#v", args)
	}
	if args[1] != "-c" {
		t.Fatalf("expected -c, got: %#v", args)
	}
	if !strings.Contains(args[2], "export A=") {
		t.Fatalf("expected env export prefix, got: %q", args[2])
	}
}

func TestAgentRunShell_CrossOS(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping cross-OS test on Windows")
	}
	ctx := context.Background()
	cmd := "echo ok"
	_, _, _, err := runpkg.RunShell(ctx, `C:\Windows\System32\cmd.exe`, "command", &cmd, nil, nil, nil, false, nil, nil)
	if err == nil {
		t.Fatal("expected error when using Windows path on non-Windows OS")
	}
	if !strings.Contains(err.Error(), "appears to be a Windows path") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
