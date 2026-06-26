package agentrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultShellPath(goos string) string {
	if goos == "windows" {
		systemRoot := os.Getenv("SystemRoot")
		if systemRoot == "" {
			systemRoot = `C:\Windows`
		}
		return filepath.Join(systemRoot, `System32\WindowsPowerShell\v1.0\powershell.exe`)
	}
	return "/bin/sh"
}

func ShellCommandInvocationArgs(goos string, shellPath string, cmdText string, env map[string]string) []string {
	if goos == "windows" {
		base := strings.ToLower(filepath.Base(shellPath))
		switch base {
		case "cmd", "cmd.exe":
			return []string{shellPath, "/C", cmdText}
		default:
			return []string{shellPath, "-NoProfile", "-NonInteractive", "-Command", cmdText}
		}
	}

	envPrefix := ""
	for k, v := range env {
		if k == "" {
			continue
		}
		envPrefix += fmt.Sprintf("export %s=%s; ", k, shellEscape(v))
	}
	return []string{shellPath, "-c", envPrefix + cmdText}
}

func RunShell(ctx context.Context, shellPath string, runMode string, command *string, script *string, workingDir *string, env map[string]string, sudo bool, runAsUser *string, sudoPassword *string) (exitCode int, stdout []byte, stderr []byte, err error) {
	if strings.TrimSpace(shellPath) == "" {
		shellPath = DefaultShellPath(runtime.GOOS)
	}

	if runtime.GOOS != "windows" && strings.Contains(shellPath, "\\") {
		if strings.HasPrefix(strings.ToUpper(shellPath), "C:\\") || strings.Contains(strings.ToLower(shellPath), "\\windows\\system32\\") {
			return -1, nil, nil, fmt.Errorf("shell path %q appears to be a Windows path, but this agent is running on %s", shellPath, runtime.GOOS)
		}
	}

	var args []string

	if runtime.GOOS == "windows" {
		if sudo || (runAsUser != nil && strings.TrimSpace(*runAsUser) != "") {
			return -1, nil, nil, errors.New("sudo/runAsUser is not supported on Windows agents")
		}
	}

	if sudo || (runAsUser != nil && strings.TrimSpace(*runAsUser) != "") {
		args = append(args, "sudo")
		if sudoPassword != nil && *sudoPassword != "" {
			args = append(args, "-S")
		} else {
			args = append(args, "-n")
		}
		if runAsUser != nil && strings.TrimSpace(*runAsUser) != "" {
			args = append(args, "-u", strings.TrimSpace(*runAsUser))
		}
		args = append(args, "--")
	}

	if strings.ToLower(strings.TrimSpace(runMode)) == "script" {
		if runtime.GOOS == "windows" {
			dir := os.TempDir()
			f, err := os.CreateTemp(dir, "cxsh-*.ps1")
			if err != nil {
				return -1, nil, []byte(err.Error()), err
			}
			name := f.Name()
			if script != nil {
				_, _ = f.WriteString(*script)
			}
			_ = f.Close()
			defer func() { _ = os.Remove(name) }()

			args = append(args, shellPath, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", name)
			cmd := exec.CommandContext(ctx, args[0], args[1:]...)
			if workingDir != nil && *workingDir != "" {
				cmd.Dir = *workingDir
			}
			envList := os.Environ()
			for k, v := range env {
				if k == "" {
					continue
				}
				envList = append(envList, fmt.Sprintf("%s=%s", k, v))
			}
			cmd.Env = envList
			var outBuf, errBuf bytes.Buffer
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			runErr := cmd.Run()
			code := 0
			if runErr != nil {
				if ee, ok := runErr.(*exec.ExitError); ok {
					code = ee.ExitCode()
				} else if ctx.Err() != nil {
					code = 124
				} else {
					code = -1
				}
			}
			return code, outBuf.Bytes(), errBuf.Bytes(), runErr
		}

		args = append(args, shellPath)
		cmd := exec.CommandContext(ctx, args[0], args[1:]...)
		if workingDir != nil && *workingDir != "" {
			cmd.Dir = *workingDir
		}
		envList := os.Environ()
		for k, v := range env {
			if k == "" {
				continue
			}
			envList = append(envList, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = envList
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		var sText strings.Builder
		if sudo && sudoPassword != nil && *sudoPassword != "" {
			sText.WriteString(*sudoPassword + "\n")
		}

		for k, v := range env {
			if k == "" {
				continue
			}
			_, _ = fmt.Fprintf(&sText, "export %s=%s\n", k, shellEscape(v))
		}

		if script != nil {
			sText.WriteString(*script)
		}

		cmd.Stdin = bytes.NewBufferString(sText.String())
		runErr := cmd.Run()
		code := 0
		if runErr != nil {
			if ee, ok := runErr.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else if ctx.Err() != nil {
				code = 124
			} else {
				code = -1
			}
		}
		return code, outBuf.Bytes(), errBuf.Bytes(), runErr
	}

	cmdText := ""
	if command != nil {
		cmdText = *command
	}

	args = append(args, ShellCommandInvocationArgs(runtime.GOOS, shellPath, cmdText, env)...)
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if workingDir != nil && *workingDir != "" {
		cmd.Dir = *workingDir
	}
	envList := os.Environ()
	for k, v := range env {
		if k == "" {
			continue
		}
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = envList
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	if sudo && sudoPassword != nil && *sudoPassword != "" {
		cmd.Stdin = strings.NewReader(*sudoPassword + "\n")
	}
	runErr := cmd.Run()
	code := 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else if ctx.Err() != nil {
			code = 124
		} else {
			code = -1
		}
	}
	return code, outBuf.Bytes(), errBuf.Bytes(), runErr
}
