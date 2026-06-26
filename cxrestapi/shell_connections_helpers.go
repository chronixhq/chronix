package cxrestapi

import (
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/internal/sshutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func decryptShellConnection(sc *models.ShellConnection) *models.ShellConnection {
	if sc == nil {
		return nil
	}
	res := *sc
	res.SSHPassword, _ = secret.DecryptPtr(sc.SSHPassword)
	res.SSHPrivateKey, _ = secret.DecryptPtr(sc.SSHPrivateKey)
	res.SSHKeyPass, _ = secret.DecryptPtr(sc.SSHKeyPass)
	res.SudoPassword, _ = secret.DecryptPtr(sc.SudoPassword)
	return &res
}

func redactShellConn(r *models.ShellConnection) gin.H {
	return gin.H{
		"id":                          r.ID,
		"name":                        r.Name,
		"description":                 r.Description,
		"agent_uuid":                  r.AgentUUID,
		"mode":                        r.Mode,
		"run_as_user":                 r.RunAsUser,
		"sudo":                        r.Sudo,
		"host":                        r.Host,
		"port":                        r.Port,
		"ssh_username":                r.SSHUsername,
		"auth_method":                 r.AuthMethod,
		"ssh_password":                redactIfSet(r.SSHPassword),
		"ssh_private_key":             redactIfSet(r.SSHPrivateKey),
		"ssh_key_pass":                redactIfSet(r.SSHKeyPass),
		"sudo_password":               redactIfSet(r.SudoPassword),
		"auto_check_enabled":          r.AutoCheckEnabled,
		"auto_check_interval_seconds": r.AutoCheckIntervalSeconds,
		"alert_emails":                r.AlertEmails,
		"alert_phones":                r.AlertPhones,
		"notify_on_failure":           r.NotifyOnFailure,
		"enabled":                     r.Enabled,
		"suspended":                   r.Suspended,
		"createdAt":                   r.CreatedAt,
		"updatedAt":                   r.UpdatedAt,
		"lastStatus":                  r.LastStatus,
		"lastError":                   r.LastError,
		"lastCheckedAt":               r.LastCheckedAt,
	}
}

func redactIfSet(p *string) any {
	if p == nil || *p == "" {
		return nil
	}
	return "<redacted>"
}

func generateKeyPair(c *gin.Context) {
	var p struct {
		Format string `json:"format"`
	}
	_ = c.ShouldBindJSON(&p)

	kp, err := sshutil.GenerateED25519KeyPair(p.Format)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to generate key pair", err.Error())
		return
	}
	restresponse.RestSuccess(c, kp)
}
