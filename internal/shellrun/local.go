package shellrun

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunLocal executes a shell step on the local host.
// runMode: "command" or "script"
// command: used when runMode=="command"
// script: used when runMode=="script"
// workingDir: optional
// env: environment variables to set/override
// sudo: when true, prepend sudo; runAsUser optionally sets -u <user>; sudoPassword provides the password for -S
func RunLocal(ctx context.Context, shellPath string, runMode string, command *string, script *string, workingDir *string, env map[string]string, sudo bool, runAsUser *string, sudoPassword *string) (exitCode int, stdout []byte, stderr []byte, err error) {
	// Build argv
	args := []string{}
	if sudo || (runAsUser != nil && strings.TrimSpace(*runAsUser) != "") {
		args = append(args, "sudo")
		if sudoPassword != nil && *sudoPassword != "" {
			// Read password from stdin
			args = append(args, "-S")
		} else {
			// Non-interactive; do not prompt
			args = append(args, "-n")
		}
		if runAsUser != nil && strings.TrimSpace(*runAsUser) != "" {
			args = append(args, "-u", strings.TrimSpace(*runAsUser))
		}
		args = append(args, "--")
	}
	// Build the shell invocation
	if runMode == "script" {
		// Write script to temp file for execution to preserve newlines
		dir := os.TempDir()
		f, err := os.CreateTemp(dir, "cxsh-*.sh")
		if err != nil {
			return -1, nil, []byte(err.Error()), err
		}
		name := f.Name()

		// Prepend environment variables to the script
		for k, v := range env {
			if k == "" {
				continue
			}
			_, _ = fmt.Fprintf(f, "export %s=%s\n", k, shellEscape(v))
		}

		if script != nil {
			_, _ = f.WriteString(*script)
		}
		_ = f.Close()
		defer func() { _ = os.Remove(name) }()
		// Use: shellPath -c '"$0"' <scriptfile> or simply execute the file with the shell
		// Simpler: run shellPath scriptfile
		args = append(args, shellPath, name)
	} else {
		// command mode (default)
		cmdText := ""
		if command != nil {
			cmdText = *command
		}

		// Prepend environment variables to the command
		envPrefix := ""
		for k, v := range env {
			if k == "" {
				continue
			}
			envPrefix += fmt.Sprintf("export %s=%s; ", k, shellEscape(v))
		}

		args = append(args, shellPath, "-c", envPrefix+cmdText)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if workingDir != nil && *workingDir != "" {
		// Ensure directory exists (best-effort)
		_ = os.MkdirAll(*workingDir, 0o755)
		cmd.Dir = *workingDir
	}
	// Environment
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
		// If context deadline exceeded/canceled, map to 124 like timeout(1)
		if ctx.Err() != nil {
			code = 124
		} else if ee, ok := runErr.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = -1
		}
	}
	// If we used a temp script file and a relative workingDir, try to clean up path from messages
	_ = filepath.Base(shellPath)
	return code, outBuf.Bytes(), errBuf.Bytes(), runErr
}
