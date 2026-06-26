package connhealth

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	eventspkg "chronix/internal/events"
	notifypkg "chronix/internal/notify"
	"chronix/internal/secret"
	"chronix/internal/shellrun"
	"chronix/internal/sqlrunner"
	"chronix/internal/webtaskrun"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/dan-sherwin/go-utilities"
)

// Start launches a lightweight background worker that periodically performs
// health checks for enabled connections with auto-check enabled.
//
// Behavior:
//   - Every sweepInterval, it loads all enabled connections with auto_check_enabled=1.
//   - For each, if last_checked_at + interval <= now, it performs a quick probe
//     using the configured driver and DSN (local or via agent when set) for databases,
//     or a quick shell command for shell connections.
//   - Updates last_status ("ok"/"error"), last_error, last_checked_at.
//   - Best-effort; errors are logged and do not terminate the loop.
func Start(ctx context.Context) {
	go runLoop(ctx)
}

const (
	sweepInterval = 15 * time.Second
	probeTimeout  = 5 * time.Second
	maxErrLen     = 600
)

func runLoop(ctx context.Context) {
	log := slog.With("component", "connhealth")
	t := time.NewTicker(sweepInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Info("connection health worker stopped")
			return
		case <-t.C:
			sweepOnce(ctx, log)
		}
	}
}

func sweepOnce(ctx context.Context, log *slog.Logger) {
	_ = sweepDatabases(ctx, log)
	_ = sweepShells(ctx, log)
	_ = sweepWebtasks(ctx, log)
}

func sweepDatabases(ctx context.Context, log *slog.Logger) error {
	rows, err := db.DbConnection.Where(db.DbConnection.AutoCheckEnabled.Neq(0), db.DbConnection.Enabled.Is(true)).Find()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, r := range rows {
		interval := time.Duration(pickInt64(r.AutoCheckIntervalSeconds, 3600)) * time.Second
		if interval <= 0 {
			interval = 5 * time.Minute // fallback
		}
		if r.LastCheckedAt != nil && r.LastCheckedAt.Add(interval).After(now) {
			continue // not yet due
		}
		status, msg := probeDatabaseOnce(ctx, r)
		prev := strings.ToLower(strings.TrimSpace(utilities.PtrVal(r.LastStatus)))
		// Update DB best-effort
		var lastErr *string
		if status != "ok" {
			m := msg
			if len(m) > maxErrLen {
				m = m[:maxErrLen]
			}
			lastErr = &m
		}
		lc := now
		updates := map[string]any{
			"last_status":     status,
			"last_error":      lastErr,
			"last_checked_at": &lc,
		}
		if _, err := db.DbConnection.Where(db.DbConnection.ID.Eq(*r.ID)).Updates(updates); err != nil {
			log.Warn("update last status", "id", r.ID, "error", err)
		} else {
			// Emit alerts on Ok↔Error transitions
			cur := strings.ToLower(strings.TrimSpace(status))
			if prev != cur && prev != "" {
				data := map[string]any{"connection_id": r.ID, "connection_name": r.Name, "kind": "database", "driver": r.Driver, "agent_uuid": r.AgentUUID, "message": msg}
				if prev == "ok" && cur == "error" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeverityError, fmt.Sprintf("Database connection '%s' became unhealthy", r.Name), nil, &data)
				} else if prev == "error" && cur == "ok" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeveritySuccess, fmt.Sprintf("Database connection '%s' recovered", r.Name), nil, &data)
				}
			}
			// Broadcast SSE to update UI lists live
			_ = eventspkg.BroadcastEvent(eventspkg.SSEEventConnectionHealth, map[string]any{
				"id":            r.ID,
				"kind":          "database",
				"lastStatus":    status,
				"lastError":     lastErr,
				"lastCheckedAt": lc,
			})
		}
	}
	return nil
}

func sweepShells(ctx context.Context, log *slog.Logger) error {
	rows, err := db.ShellConnection.Where(db.ShellConnection.AutoCheckEnabled.Neq(0), db.ShellConnection.Enabled.Is(true)).Find()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, r := range rows {
		interval := time.Duration(r.AutoCheckIntervalSeconds) * time.Second
		if interval <= 0 {
			interval = 5 * time.Minute // fallback
		}
		if r.LastCheckedAt != nil && r.LastCheckedAt.Add(interval).After(now) {
			continue // not yet due
		}
		status, msg := probeShellOnce(ctx, r)
		prev := strings.ToLower(strings.TrimSpace(utilities.PtrVal(r.LastStatus)))
		// Update DB best-effort
		var lastErr *string
		if status != "ok" {
			m := msg
			if len(m) > maxErrLen {
				m = m[:maxErrLen]
			}
			lastErr = &m
		}
		lc := now
		updates := map[string]any{
			"last_status":     status,
			"last_error":      lastErr,
			"last_checked_at": &lc,
		}
		if _, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(*r.ID)).Updates(updates); err != nil {
			log.Warn("update last status", "id", r.ID, "error", err)
		} else {
			// Emit alerts on Ok↔Error transitions
			cur := strings.ToLower(strings.TrimSpace(status))
			if prev != cur && prev != "" {
				data := map[string]any{"connection_id": r.ID, "connection_name": r.Name, "kind": "shell", "mode": r.Mode, "agent_uuid": r.AgentUUID, "message": msg}
				if prev == "ok" && cur == "error" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeverityError, fmt.Sprintf("Shell connection '%s' became unhealthy", r.Name), nil, &data)
				} else if prev == "error" && cur == "ok" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeveritySuccess, fmt.Sprintf("Shell connection '%s' recovered", r.Name), nil, &data)
				}
			}
			// Broadcast SSE to update UI lists live
			_ = eventspkg.BroadcastEvent(eventspkg.SSEEventConnectionHealth, map[string]any{
				"id":            r.ID,
				"kind":          "shell",
				"lastStatus":    status,
				"lastError":     lastErr,
				"lastCheckedAt": lc,
			})
		}
	}
	return nil
}

func probeDatabaseOnce(parent context.Context, r *models.DbConnection) (string, string) {
	dsn := ""
	if r.Dsn != nil {
		dsn, _ = secret.Decrypt(*r.Dsn)
	}
	if strings.TrimSpace(dsn) == "" {
		return "error", "empty DSN"
	}
	drv := strings.ToLower(strings.TrimSpace(r.Driver))
	var runner sqlrunner.SQLRunner
	if r.AgentUUID != nil && strings.TrimSpace(*r.AgentUUID) != "" {
		runner = sqlrunner.AgentRunner{AgentID: strings.TrimSpace(*r.AgentUUID)}
	} else {
		runner = sqlrunner.LocalRunner{}
	}
	ctx, cancel := context.WithTimeout(parent, probeTimeout)
	defer cancel()
	var probe string
	switch drv {
	case "postgres", "pgx":
		probe = "SELECT 1"
	case "sqlite":
		probe = "SELECT 1"
	default:
		return "error", "unsupported driver: " + drv
	}
	rows, _, err := runner.Query(ctx, drv, dsn, probe, nil, 1)
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "agent not connected") {
			return "error", "agent not connected"
		}
		return "error", msg
	}
	if rows >= 0 {
		return "ok", "connection ok"
	}
	return "ok", "connection ok"
}

func probeShellOnce(parent context.Context, r *models.ShellConnection) (string, string) {
	ctx, cancel := context.WithTimeout(parent, probeTimeout)
	defer cancel()

	// Decrypt secrets
	pwd, _ := secret.DecryptPtr(r.SSHPassword)
	pk, _ := secret.DecryptPtr(r.SSHPrivateKey)
	kp, _ := secret.DecryptPtr(r.SSHKeyPass)
	sp, _ := secret.DecryptPtr(r.SudoPassword)

	// Copy connection and use decrypted secrets
	sc := *r
	sc.SSHPassword = pwd
	sc.SSHPrivateKey = pk
	sc.SSHKeyPass = kp
	sc.SudoPassword = sp

	mode := strings.ToLower(sc.Mode)
	agentID := ""
	if sc.AgentUUID != nil {
		agentID = strings.TrimSpace(*sc.AgentUUID)
	}

	cmd := "echo ok"
	shellPath := "/bin/sh"
	if agentID != "" && (mode == "localhost" || mode == "") {
		// Let the agent pick an OS-appropriate default shell (Windows agents won’t have /bin/sh).
		shellPath = ""
	}

	var (
		exitCode int
		err      error
	)

	if agentID != "" {
		var sshCfg *shellrun.SSHConfig
		if mode == "ssh" {
			pVal := int(pickInt64(sc.Port, 22))
			sshCfg = &shellrun.SSHConfig{
				Mode:          "ssh",
				Host:          sc.Host,
				Port:          &pVal,
				Username:      sc.SSHUsername,
				AuthMethod:    sc.AuthMethod,
				Password:      sc.SSHPassword,
				PrivateKey:    sc.SSHPrivateKey,
				KeyPassphrase: sc.SSHKeyPass,
				SudoPassword:  sc.SudoPassword,
			}
		}
		exitCode, _, _, err = shellrun.RunAgent(ctx, agentID, shellPath, "command", &cmd, nil, nil, nil, sshCfg, pickBool(sc.Sudo), sc.RunAsUser, sc.SudoPassword)
	} else {
		if mode == "localhost" || mode == "" {
			exitCode, _, _, err = shellrun.RunLocal(ctx, shellPath, "command", &cmd, nil, nil, nil, pickBool(sc.Sudo), sc.RunAsUser, sc.SudoPassword)
		} else {
			exitCode, _, _, err = shellrun.RunSSH(ctx, &sc, shellPath, "command", &cmd, nil, nil, nil)
		}
	}

	if err != nil {
		return "error", err.Error()
	}
	if exitCode != 0 {
		return "error", fmt.Sprintf("exit code %d", exitCode)
	}
	return "ok", "connection ok"
}

func pickInt64(p *int64, def int64) int64 {
	if p == nil {
		return def
	}
	return *p
}

func pickBool(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func sweepWebtasks(ctx context.Context, log *slog.Logger) error {
	rows, err := db.WebtaskConnection.Where(db.WebtaskConnection.AutoCheckEnabled.Neq(0), db.WebtaskConnection.Enabled.Is(true)).Find()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, r := range rows {
		interval := time.Duration(pickInt64(r.AutoCheckIntervalSeconds, 3600)) * time.Second
		if interval <= 0 {
			interval = 5 * time.Minute // fallback
		}
		if r.LastCheckedAt != nil && r.LastCheckedAt.Add(interval).After(now) {
			continue // not yet due
		}
		status, msg := probeWebtaskOnce(ctx, r)
		prev := strings.ToLower(strings.TrimSpace(utilities.PtrVal(r.LastStatus)))
		// Update DB best-effort
		var lastErr *string
		if status != "ok" {
			m := msg
			if len(m) > maxErrLen {
				m = m[:maxErrLen]
			}
			lastErr = &m
		}
		lc := now
		updates := map[string]any{
			"last_status":     status,
			"last_error":      lastErr,
			"last_checked_at": &lc,
		}
		if _, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(*r.ID)).Updates(updates); err != nil {
			log.Warn("update last status", "id", r.ID, "error", err)
		} else {
			// Emit alerts on Ok↔Error transitions
			cur := strings.ToLower(strings.TrimSpace(status))
			if prev != cur && prev != "" {
				data := map[string]any{"connection_id": r.ID, "connection_name": r.Name, "kind": "webtask", "base_url": r.BaseURL, "agent_uuid": r.AgentUUID, "message": msg}
				if prev == "ok" && cur == "error" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeverityError, fmt.Sprintf("Web Task connection '%s' became unhealthy", r.Name), nil, &data)
				} else if prev == "error" && cur == "ok" {
					notifypkg.TryCreateNotification(notifypkg.CategoryConnection, notifypkg.SeveritySuccess, fmt.Sprintf("Web Task connection '%s' recovered", r.Name), nil, &data)
				}
			}
			// Broadcast SSE to update UI lists live
			_ = eventspkg.BroadcastEvent(eventspkg.SSEEventConnectionHealth, map[string]any{
				"id":            r.ID,
				"kind":          "webtask",
				"lastStatus":    status,
				"lastError":     lastErr,
				"lastCheckedAt": lc,
			})
		}
	}
	return nil
}

func probeWebtaskOnce(parent context.Context, r *models.WebtaskConnection) (string, string) {
	ctx, cancel := context.WithTimeout(parent, probeTimeout)
	defer cancel()

	var runner webtaskrun.WebTaskRunner
	if r.AgentUUID != nil && strings.TrimSpace(*r.AgentUUID) != "" {
		runner = &webtaskrun.AgentRunner{AgentID: strings.TrimSpace(*r.AgentUUID)}
	} else {
		runner = webtaskrun.NewLocalRunner()
	}

	// Simple GET probe to base URL
	url := utilities.PtrVal(r.BaseURL)
	if url == "" {
		return "error", "empty base URL"
	}

	res, err := runner.Execute(ctx, "GET", url, nil, nil, probeTimeout)
	if err != nil {
		return "error", err.Error()
	}

	if res.StatusCode >= 200 && res.StatusCode < 400 {
		return "ok", "connection ok"
	}
	return "error", fmt.Sprintf("HTTP status %d", res.StatusCode)
}
