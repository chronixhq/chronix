package agentrun

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host          string
	Port          int
	Username      string
	AuthMethod    string
	Password      string
	PrivateKey    string
	KeyPassphrase string
	SudoPassword  string
}

func RunSSH(ctx context.Context, sshCfg *SSHConfig, shellPath string, runMode string, command *string, script *string, workingDir *string, env map[string]string, sudo bool, runAsUser *string) (exitCode int, stdout []byte, stderr []byte, err error) {
	if sshCfg == nil || sshCfg.Host == "" || sshCfg.Username == "" {
		return -1, nil, nil, errors.New("invalid ssh configuration")
	}
	host := strings.TrimSpace(sshCfg.Host)
	user := strings.TrimSpace(sshCfg.Username)
	port := sshCfg.Port
	if port <= 0 {
		port = 22
	}

	var auth []ssh.AuthMethod
	if strings.ToLower(sshCfg.AuthMethod) == "password" {
		auth = append(auth, ssh.Password(sshCfg.Password))
	} else {
		if sshCfg.PrivateKey == "" {
			return -1, nil, nil, errors.New("key auth selected but private key is empty")
		}
		var signer ssh.Signer
		var perr error
		if sshCfg.KeyPassphrase != "" {
			signer, perr = ssh.ParsePrivateKeyWithPassphrase([]byte(sshCfg.PrivateKey), []byte(sshCfg.KeyPassphrase))
		} else {
			signer, perr = ssh.ParsePrivateKey([]byte(sshCfg.PrivateKey))
		}
		if perr != nil {
			return -1, nil, nil, fmt.Errorf("parse private key: %w", perr)
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

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			d.Close()
		case <-done:
		}
	}()
	defer close(done)

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

	for k, v := range env {
		_ = sess.Setenv(k, v)
	}

	var outBuf, errBuf bytes.Buffer
	sess.Stdout = &outBuf
	sess.Stderr = &errBuf

	cmd := ""
	for k, v := range env {
		if k == "" {
			continue
		}
		cmd += fmt.Sprintf("export %s=%s; ", k, shellEscape(v))
	}

	if workingDir != nil && *workingDir != "" {
		cmd += fmt.Sprintf("cd %s; ", shellEscape(*workingDir))
	}

	rau := ""
	if runAsUser != nil {
		rau = strings.TrimSpace(*runAsUser)
	}
	if sudo || rau != "" {
		if sudo && sshCfg.SudoPassword != "" {
			cmd += "sudo -S "
		} else {
			cmd += "sudo -n "
		}
		if rau != "" {
			cmd += fmt.Sprintf("-u %s ", shellEscape(rau))
		}
		cmd += "-- "
	}

	if runMode == "script" {
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

	if sudo && sshCfg.SudoPassword != "" {
		sess.Stdin = strings.NewReader(sshCfg.SudoPassword + "\n")
	}

	runErr := sess.Run(cmd)
	code := 0
	if runErr != nil {
		if ee, ok := runErr.(*ssh.ExitError); ok {
			code = ee.ExitStatus()
		} else if ctx.Err() != nil {
			code = 124
		} else {
			code = -1
		}
	}
	return code, outBuf.Bytes(), errBuf.Bytes(), runErr
}

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
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func shellQuote(s string) string { return shellEscape(s) }
