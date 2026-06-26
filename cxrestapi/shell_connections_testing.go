package cxrestapi

import (
	"bytes"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/internal/shellrun"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
)

func testShellConnection(c *gin.Context) {
	id := atoi64(c.Param("id"))
	rec, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Shell connection not found")
		return
	}

	dec := decryptShellConnection(rec)
	mode := strings.ToLower(strings.TrimSpace(dec.Mode))
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	var ok bool
	var msg string

	if dec.AgentUUID != nil && strings.TrimSpace(*dec.AgentUUID) != "" {
		ok, msg = testAgentShellConnection(ctx, dec, mode)
	} else if mode == "localhost" || mode == "" {
		ok, msg = testLocalShellConnection(ctx, dec)
	} else {
		if err := trySSHCommand(dec, "echo ok"); err != nil {
			ok = false
			msg = fmt.Sprintf("ssh test failed: %v", err)
		} else {
			ok = true
			msg = "ssh connectivity ok"
		}
	}

	now := time.Now().UTC()
	status := "ok"
	var lastErr *string
	if !ok {
		status = "error"
		lastErr = &msg
	}
	rec.LastStatus = &status
	rec.LastError = lastErr
	rec.LastCheckedAt = &now
	_ = db.ShellConnection.Save(rec)
	restresponse.RestSuccess(c, gin.H{"ok": ok, "message": msg})
}

func testShellConnectionFromDraft(c *gin.Context) {
	var p shellConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := validateShellConnPayload(&p, p.ID != nil && *p.ID > 0); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}

	if p.ID != nil && *p.ID > 0 {
		rec, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(*p.ID)).First()
		if err == nil && rec != nil {
			if (p.SSHPassword == nil || *p.SSHPassword == "<redacted>") && rec.SSHPassword != nil {
				p.SSHPassword = rec.SSHPassword
			}
			if (p.SSHPrivateKey == nil || *p.SSHPrivateKey == "<redacted>") && rec.SSHPrivateKey != nil {
				p.SSHPrivateKey = rec.SSHPrivateKey
			}
			if (p.SSHKeyPass == nil || *p.SSHKeyPass == "<redacted>") && rec.SSHKeyPass != nil {
				p.SSHKeyPass = rec.SSHKeyPass
			}
			if (p.SudoPassword == nil || *p.SudoPassword == "<redacted>") && rec.SudoPassword != nil {
				p.SudoPassword = rec.SudoPassword
			}
		}
	}

	p.SSHPassword, _ = secret.DecryptPtr(p.SSHPassword)
	p.SSHPrivateKey, _ = secret.DecryptPtr(p.SSHPrivateKey)
	p.SSHKeyPass, _ = secret.DecryptPtr(p.SSHKeyPass)
	p.SudoPassword, _ = secret.DecryptPtr(p.SudoPassword)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()
	mode := ""
	if p.Mode != nil {
		mode = strings.ToLower(strings.TrimSpace(*p.Mode))
	}
	var ok bool
	var msg string

	if p.AgentUUID != nil && strings.TrimSpace(*p.AgentUUID) != "" {
		tmp := draftPayloadToShellConnection(&p)
		ok, msg = testAgentShellConnection(ctx, tmp, mode)
	} else if mode == "localhost" || mode == "" {
		tmp := draftPayloadToShellConnection(&p)
		ok, msg = testLocalShellConnection(ctx, tmp)
	} else {
		tmp := draftPayloadToShellConnection(&p)
		if err := trySSHCommand(tmp, "echo ok"); err != nil {
			ok = false
			msg = fmt.Sprintf("ssh test failed: %v", err)
		} else {
			ok = true
			msg = "ssh connectivity ok"
		}
	}
	restresponse.RestSuccess(c, gin.H{"ok": ok, "message": msg})
}

func testAgentShellConnection(ctx context.Context, sc *models.ShellConnection, mode string) (bool, string) {
	agentID := strings.TrimSpace(utilities.PtrVal(sc.AgentUUID))
	cmdText := "echo ok"
	var sshCfg *shellrun.SSHConfig
	if mode == "ssh" {
		port := int(pickInt64Val(sc.Port, 22))
		sshCfg = &shellrun.SSHConfig{
			Mode:          "ssh",
			Host:          sc.Host,
			Port:          &port,
			Username:      sc.SSHUsername,
			AuthMethod:    sc.AuthMethod,
			Password:      sc.SSHPassword,
			PrivateKey:    sc.SSHPrivateKey,
			KeyPassphrase: sc.SSHKeyPass,
			SudoPassword:  sc.SudoPassword,
		}
	}
	shellPath := "/bin/sh"
	if mode == "localhost" || mode == "" {
		shellPath = ""
	}
	exitCode, _, stderr, runErr := shellrun.RunAgent(ctx, agentID, shellPath, "command", &cmdText, nil, nil, nil, sshCfg, pickBool(sc.Sudo, false), sc.RunAsUser, sc.SudoPassword)
	if runErr != nil || exitCode != 0 {
		if runErr != nil {
			msg := fmt.Sprintf("agent shell test failed: %v", runErr)
			if len(stderr) > 0 {
				msg += ": " + strings.TrimSpace(string(stderr))
			}
			return false, msg
		}
		if len(stderr) > 0 {
			return false, fmt.Sprintf("agent shell test failed: exit %d: %s", exitCode, strings.TrimSpace(string(stderr)))
		}
		return false, fmt.Sprintf("agent shell test failed: exit %d", exitCode)
	}
	return true, "agent connectivity ok"
}

func testLocalShellConnection(ctx context.Context, sc *models.ShellConnection) (bool, string) {
	shellPath := "/bin/sh"
	cmdText := "echo ok"
	exitCode, _, stderr, runErr := shellrun.RunLocal(ctx, shellPath, "command", &cmdText, nil, nil, map[string]string{}, pickBool(sc.Sudo, false), sc.RunAsUser, sc.SudoPassword)
	if runErr != nil || exitCode != 0 {
		if runErr != nil {
			msg := fmt.Sprintf("local shell test failed: %v", runErr)
			if len(stderr) > 0 {
				msg += ": " + strings.TrimSpace(string(stderr))
			}
			return false, msg
		}
		if len(stderr) > 0 {
			return false, fmt.Sprintf("local shell test failed: exit %d: %s", exitCode, strings.TrimSpace(string(stderr)))
		}
		return false, fmt.Sprintf("local shell test failed: exit %d", exitCode)
	}
	return true, "localhost connectivity ok"
}

func trySSHCommand(sc *models.ShellConnection, cmd string) error {
	sc = decryptShellConnection(sc)
	host := utilities.PtrVal(sc.Host)
	port := int64(22)
	if sc.Port != nil && *sc.Port > 0 {
		port = *sc.Port
	}
	user := utilities.PtrVal(sc.SSHUsername)
	if user == "" || host == "" {
		return errors.New("invalid ssh parameters")
	}
	var auth []ssh.AuthMethod
	if sc.AuthMethod != nil && strings.ToLower(*sc.AuthMethod) == "password" {
		if sc.SSHPassword == nil {
			return errors.New("password auth selected but password is empty")
		}
		auth = append(auth, ssh.Password(*sc.SSHPassword))
	} else {
		if sc.SSHPrivateKey == nil {
			return errors.New("key auth selected but private key is empty")
		}
		var signer ssh.Signer
		var err error
		if sc.SSHKeyPass != nil && *sc.SSHKeyPass != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(*sc.SSHPrivateKey), []byte(*sc.SSHKeyPass))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(*sc.SSHPrivateKey))
		}
		if err != nil {
			return fmt.Errorf("parse private key: %w", err)
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
	slog.Info("ssh dialing", "component", "cxrestapi", "op", "shell-test", "host", host, "port", port)
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	sess, err := client.NewSession()
	if err != nil {
		return err
	}
	defer func() { _ = sess.Close() }()

	var stderr bytes.Buffer
	sess.Stderr = &stderr

	if cmd == "" {
		cmd = "echo ok"
	}

	sudo := false
	if sc.Sudo != nil {
		sudo = *sc.Sudo
	}
	runAsUser := strings.TrimSpace(utilities.PtrVal(sc.RunAsUser))

	finalCmd := ""
	if sudo || runAsUser != "" {
		if sudo && sc.SudoPassword != nil && *sc.SudoPassword != "" {
			finalCmd = "sudo -S "
		} else {
			finalCmd = "sudo -n "
		}
		if runAsUser != "" {
			finalCmd += fmt.Sprintf("-u '%s' ", strings.ReplaceAll(runAsUser, "'", "'\\''"))
		}
		finalCmd += "-- "
	}
	finalCmd += cmd

	if sudo && sc.SudoPassword != nil && *sc.SudoPassword != "" {
		sess.Stdin = strings.NewReader(*sc.SudoPassword + "\n")
	}

	if err := sess.Run(finalCmd); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return err
	}
	return nil
}

func draftPayloadToShellConnection(p *shellConnPayload) *models.ShellConnection {
	sudoVal := pickBool(p.Sudo, false)
	now := time.Now().UTC()
	name := utilities.PtrVal(p.Name)
	mode := utilities.PtrVal(p.Mode)
	return &models.ShellConnection{
		Name:          name,
		Description:   p.Description,
		AgentUUID:     p.AgentUUID,
		Mode:          mode,
		RunAsUser:     p.RunAsUser,
		Sudo:          &sudoVal,
		Host:          p.Host,
		Port:          p.Port,
		SSHUsername:   p.SSHUsername,
		AuthMethod:    p.AuthMethod,
		SSHPassword:   p.SSHPassword,
		SSHPrivateKey: p.SSHPrivateKey,
		SSHKeyPass:    p.SSHKeyPass,
		SudoPassword:  p.SudoPassword,
		UpdatedAt:     now,
	}
}
