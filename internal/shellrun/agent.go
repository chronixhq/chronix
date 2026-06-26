package shellrun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"chronix/internal/agentmux"
)

// Agent request/response contracts (mirrors cmd/chronix-agent handler)
type shellExecReq struct {
	ShellPath  string            `json:"shellPath"`
	RunMode    string            `json:"runMode"` // command | script
	Command    *string           `json:"command"`
	ScriptText *string           `json:"scriptText"`
	WorkingDir *string           `json:"workingDir"`
	Env        map[string]string `json:"env"`
	TimeoutMs  int64             `json:"timeoutMs"`

	Sudo         bool    `json:"sudo,omitempty"`
	RunAsUser    *string `json:"runAsUser,omitempty"`
	SudoPassword *string `json:"sudoPassword,omitempty"`

	// SSH details for agent-to-remote SSH execution
	Mode          string  `json:"mode,omitempty"`
	Host          *string `json:"host,omitempty"`
	Port          *int    `json:"port,omitempty"`
	SSHUsername   *string `json:"sshUsername,omitempty"`
	AuthMethod    *string `json:"authMethod,omitempty"`
	SSHPassword   *string `json:"sshPassword,omitempty"`
	SSHPrivateKey *string `json:"sshPrivateKey,omitempty"`
	SSHKeyPass    *string `json:"sshKeyPass,omitempty"`
}

type shellExecOK struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type agentError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Detail  map[string]any `json:"detail"`
}

type SSHConfig struct {
	Mode          string
	Host          *string
	Port          *int
	Username      *string
	AuthMethod    *string
	Password      *string
	PrivateKey    *string
	KeyPassphrase *string
	SudoPassword  *string
}

// RunAgent dispatches a shell exec to the given agent UUID over agentmux.
func RunAgent(ctx context.Context, agentID string, shellPath string, runMode string, command *string, script *string, workingDir *string, env map[string]string, ssh *SSHConfig, sudo bool, runAsUser *string, sudoPassword *string) (exitCode int, stdout []byte, stderr []byte, err error) {
	c := agentmux.DefaultManager.Get(agentID)
	if c == nil {
		return 0, nil, nil, errors.New("agent not connected")
	}
	to := deadlineOr(ctx, 30*time.Second)
	req := shellExecReq{
		ShellPath:    shellPath,
		RunMode:      runMode,
		Command:      command,
		ScriptText:   script,
		WorkingDir:   workingDir,
		Env:          env,
		TimeoutMs:    int64(time.Until(to).Milliseconds()),
		Sudo:         sudo,
		RunAsUser:    runAsUser,
		SudoPassword: sudoPassword,
	}
	if ssh != nil {
		req.Mode = ssh.Mode
		req.Host = ssh.Host
		req.Port = ssh.Port
		req.SSHUsername = ssh.Username
		req.AuthMethod = ssh.AuthMethod
		req.SSHPassword = ssh.Password
		req.SSHPrivateKey = ssh.PrivateKey
		req.SSHKeyPass = ssh.KeyPassphrase
		if ssh.SudoPassword != nil {
			req.SudoPassword = ssh.SudoPassword
		}
	}
	typ, payload, err := c.Request(ctx, "shell.run", req)
	if err != nil {
		return -1, nil, nil, err
	}
	switch typ {
	case "shell.run.ok":
		var ok shellExecOK
		if err := json.Unmarshal(payload, &ok); err != nil {
			return -1, nil, nil, err
		}
		return ok.ExitCode, []byte(ok.Stdout), []byte(ok.Stderr), nil
	case "shell.error":
		var e agentError
		_ = json.Unmarshal(payload, &e)
		if e.Message == "" {
			e.Message = "agent shell error"
		}
		return -1, nil, nil, fmt.Errorf("%s", e.Message)
	default:
		return -1, nil, nil, fmt.Errorf("unexpected response: %s", typ)
	}
}

func deadlineOr(ctx context.Context, def time.Duration) time.Time {
	if dl, ok := ctx.Deadline(); ok {
		return dl
	}
	return time.Now().Add(def)
}
