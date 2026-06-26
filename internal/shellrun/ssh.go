package shellrun

import (
	"bytes"
	"chronix/internal/db/models"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// RunSSH executes a shell step via SSH described by the ShellConnection.
// For simplicity, we stream stdout/stderr into buffers and return them.
func RunSSH(ctx context.Context, conn *models.ShellConnection, shellPath string, runMode string, command *string, script *string, workingDir *string, env map[string]string) (exitCode int, stdout []byte, stderr []byte, err error) {
	if conn == nil || conn.Host == nil || conn.SSHUsername == nil {
		return 1, nil, nil, errors.New("invalid ssh connection")
	}
	host := strings.TrimSpace(*conn.Host)
	user := strings.TrimSpace(*conn.SSHUsername)
	port := int64(22)
	if conn.Port != nil && *conn.Port > 0 {
		port = *conn.Port
	}
	if host == "" || user == "" {
		return 1, nil, nil, errors.New("missing host or user")
	}
	var auth []ssh.AuthMethod
	if conn.AuthMethod != nil && strings.ToLower(*conn.AuthMethod) == "password" {
		if conn.SSHPassword == nil {
			return 1, nil, nil, errors.New("password auth selected but password is empty")
		}
		auth = append(auth, ssh.Password(*conn.SSHPassword))
	} else {
		if conn.SSHPrivateKey == nil {
			return 1, nil, nil, errors.New("key auth selected but private key is empty")
		}
		var signer ssh.Signer
		var perr error
		if conn.SSHKeyPass != nil && *conn.SSHKeyPass != "" {
			signer, perr = ssh.ParsePrivateKeyWithPassphrase([]byte(*conn.SSHPrivateKey), []byte(*conn.SSHKeyPass))
		} else {
			signer, perr = ssh.ParsePrivateKey([]byte(*conn.SSHPrivateKey))
		}
		if perr != nil {
			return 1, nil, nil, fmt.Errorf("parse private key: %w", perr)
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	d := &netDialer{cfg: cfg, addr: addr}
	// Honor context cancellation by closing the underlying connection if set
	go func() {
		<-ctx.Done()
		d.Close()
	}()
	client, err := d.Dial()
	if err != nil {
		return -1, nil, nil, err
	}
	defer func() { _ = client.Close() }()
	sess, err := client.NewSession()
	if err != nil {
		return -1, nil, nil, err
	}
	defer func() { _ = sess.Close() }()
	// Set env
	for k, v := range env {
		_ = sess.Setenv(k, v)
	}
	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf
	// Compose remote command
	cmd := ""

	// Prepend environment variables since sess.Setenv is often restricted by SSH servers
	for k, v := range env {
		if k == "" {
			continue
		}
		cmd += fmt.Sprintf("export %s=%s; ", k, shellEscape(v))
	}

	if workingDir != nil && *workingDir != "" {
		cmd += fmt.Sprintf("cd %s; ", shellEscape(*workingDir))
	}

	sudo := false
	if conn.Sudo != nil {
		sudo = *conn.Sudo
	}
	runAsUser := ""
	if conn.RunAsUser != nil {
		runAsUser = strings.TrimSpace(*conn.RunAsUser)
	}
	if sudo || runAsUser != "" {
		if sudo && conn.SudoPassword != nil && *conn.SudoPassword != "" {
			cmd += "sudo -S "
		} else {
			cmd += "sudo -n "
		}
		if runAsUser != "" {
			cmd += fmt.Sprintf("-u %s ", shellEscape(runAsUser))
		}
		cmd += "-- "
	}

	if runMode == "script" {
		// Use here-doc to avoid temp files
		s := ""
		if script != nil {
			s = *script
		}
		cmd += fmt.Sprintf("%s <<'CXEOF'\n%s\nCXEOF", shellEscape(shellPath), s)
	} else {
		c := ""
		if command != nil {
			c = *command
		}
		cmd += fmt.Sprintf("%s -c %s", shellEscape(shellPath), shellQuote(c))
	}

	if sudo && conn.SudoPassword != nil && *conn.SudoPassword != "" {
		sess.Stdin = strings.NewReader(*conn.SudoPassword + "\n")
	}

	runErr := sess.Run(cmd)
	code := 0
	if runErr != nil {
		if ee, ok := runErr.(*ssh.ExitError); ok {
			code = ee.ExitStatus()
		} else {
			// Context canceled/timeout or other transport error
			if ctx.Err() != nil {
				code = 124
			} else {
				code = -1
			}
		}
	}
	return code, outBuf.Bytes(), errBuf.Bytes(), runErr
}

// netDialer wraps ssh.Dial to support closing on context cancel.
type netDialer struct {
	cfg  *ssh.ClientConfig
	addr string
	c    *ssh.Client
}

func (d *netDialer) Dial() (*ssh.Client, error) {
	c, err := ssh.Dial("tcp", d.addr, d.cfg)
	if err != nil {
		return nil, err
	}
	d.c = c
	return c, nil
}

func (d *netDialer) Close() {
	if d.c != nil {
		_ = d.c.Close()
	}
}

func shellEscape(s string) string {
	// Minimal escape: wrap in single quotes and escape existing single quotes
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellQuote(s string) string { return shellEscape(s) }
