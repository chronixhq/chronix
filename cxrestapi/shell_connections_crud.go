package cxrestapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type shellConnPayload struct {
	ID          *int64  `json:"id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	AgentUUID   *string `json:"agent_uuid"`

	Mode      *string `json:"mode"`
	RunAsUser *string `json:"run_as_user"`
	Sudo      *bool   `json:"sudo"`

	Host          *string `json:"host"`
	Port          *int64  `json:"port"`
	SSHUsername   *string `json:"ssh_username"`
	AuthMethod    *string `json:"auth_method"`
	SSHPassword   *string `json:"ssh_password"`
	SSHPrivateKey *string `json:"ssh_private_key"`
	SSHKeyPass    *string `json:"ssh_key_pass"`
	SudoPassword  *string `json:"sudo_password"`

	AutoCheckEnabled         *bool `json:"auto_check_enabled"`
	AutoCheckIntervalSeconds int64 `json:"auto_check_interval_seconds"`

	AlertEmails     *string `json:"alert_emails"`
	AlertPhones     *string `json:"alert_phones"`
	NotifyOnFailure *bool   `json:"notify_on_failure"`
	Enabled         *bool   `json:"enabled"`
	Suspended       *bool   `json:"suspended"`
}

func validateShellConnPayload(p *shellConnPayload, isUpdate bool) error {
	if !isUpdate {
		if p.Name == nil || strings.TrimSpace(*p.Name) == "" {
			return errors.New("name is required")
		}
		if p.Mode == nil || strings.TrimSpace(*p.Mode) == "" {
			return errors.New("mode is required")
		}
	} else {
		if p.Name != nil && strings.TrimSpace(*p.Name) == "" {
			return errors.New("name cannot be empty")
		}
		if p.Mode != nil && strings.TrimSpace(*p.Mode) == "" {
			return errors.New("mode cannot be empty")
		}
	}

	m := ""
	if p.Mode != nil {
		m = strings.ToLower(strings.TrimSpace(*p.Mode))
		if m != "localhost" && m != "ssh" {
			return errors.New("mode must be 'localhost' or 'ssh'")
		}
	}

	if m == "ssh" {
		if p.Host == nil || strings.TrimSpace(*p.Host) == "" {
			return errors.New("host is required for ssh mode")
		}
		if p.Port == nil || *p.Port <= 0 {
			return errors.New("port is required for ssh mode")
		}
		if p.SSHUsername == nil || strings.TrimSpace(*p.SSHUsername) == "" {
			return errors.New("ssh_username is required for ssh mode")
		}
		if p.AuthMethod == nil {
			return errors.New("auth_method is required for ssh mode")
		}
		am := strings.ToLower(strings.TrimSpace(*p.AuthMethod))
		if am != "password" && am != "key" {
			return errors.New("auth_method must be 'password' or 'key'")
		}
		if !isUpdate {
			if am == "password" {
				if p.SSHPassword == nil || *p.SSHPassword == "" {
					return errors.New("ssh_password is required for password auth")
				}
			} else if p.SSHPrivateKey == nil || *p.SSHPrivateKey == "" {
				return errors.New("ssh_private_key is required for key auth")
			}
		}
	}
	return nil
}

func listShellConnections(c *gin.Context) {
	rows, err := db.ShellConnection.Order(db.ShellConnection.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list shell connections", err.Error())
		return
	}
	uuids := make([]string, 0)
	uuidSeen := make(map[string]struct{})
	for _, it := range rows {
		if it.AgentUUID != nil {
			u := strings.TrimSpace(*it.AgentUUID)
			if u != "" {
				if _, ok := uuidSeen[u]; !ok {
					uuidSeen[u] = struct{}{}
					uuids = append(uuids, u)
				}
			}
		}
	}
	agentNames := make(map[string]string, len(uuids))
	agentOS := make(map[string]string, len(uuids))
	if len(uuids) > 0 {
		if agents, err := db.Agent.Where(db.Agent.UUID.In(uuids...)).Find(); err == nil {
			for _, a := range agents {
				agentNames[a.UUID] = a.Name
				if a.MetadataJSON != nil {
					if osVal, ok := (*a.MetadataJSON)["os"].(string); ok {
						agentOS[a.UUID] = osVal
					}
				}
			}
		}
	}

	resp := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		row := redactShellConn(r)
		if r.AgentUUID != nil {
			u := strings.TrimSpace(*r.AgentUUID)
			if name, ok := agentNames[u]; ok && name != "" {
				row["agent_name"] = name
			}
			if osVal, ok := agentOS[u]; ok && osVal != "" {
				row["agent_os"] = osVal
			}
		}
		resp = append(resp, row)
	}
	restresponse.RestSuccess(c, resp)
}

func getShellConnection(c *gin.Context) {
	id := atoi64(c.Param("id"))
	r, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(strings.ToLower(err.Error()), "record not found") {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Shell connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load shell connection", err.Error())
		return
	}
	resp := redactShellConn(r)
	if r.AgentUUID != nil {
		if agent, err := db.Agent.Where(db.Agent.UUID.Eq(strings.TrimSpace(*r.AgentUUID))).First(); err == nil {
			resp["agent_name"] = agent.Name
			if agent.MetadataJSON != nil {
				if osVal, ok := (*agent.MetadataJSON)["os"].(string); ok {
					resp["agent_os"] = osVal
				}
			}
		}
	}
	restresponse.RestSuccess(c, resp)
}

func createShellConnection(c *gin.Context) {
	var p shellConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := validateShellConnPayload(&p, false); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}

	now := time.Now().UTC()
	sudoVal := pickBool(p.Sudo, false)
	name := ""
	if p.Name != nil {
		name = *p.Name
	}
	mode := ""
	if p.Mode != nil {
		mode = *p.Mode
	}
	rec := &models.ShellConnection{
		Name:                     name,
		Description:              p.Description,
		AgentUUID:                p.AgentUUID,
		Mode:                     mode,
		RunAsUser:                p.RunAsUser,
		Sudo:                     &sudoVal,
		Host:                     p.Host,
		Port:                     p.Port,
		SSHUsername:              p.SSHUsername,
		AuthMethod:               p.AuthMethod,
		AutoCheckEnabled:         utilities.Ptr(int64(0)),
		AutoCheckIntervalSeconds: p.AutoCheckIntervalSeconds,
		AlertEmails:              p.AlertEmails,
		AlertPhones:              p.AlertPhones,
		NotifyOnFailure:          p.NotifyOnFailure,
		Enabled:                  true,
		Suspended:                false,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	if p.AutoCheckEnabled != nil && *p.AutoCheckEnabled {
		rec.AutoCheckEnabled = utilities.Ptr(int64(1))
	}
	if rec.AutoCheckIntervalSeconds <= 0 {
		rec.AutoCheckIntervalSeconds = 300
	}
	rec.SSHPassword, _ = secret.EncryptPtr(p.SSHPassword)
	rec.SSHPrivateKey, _ = secret.EncryptPtr(p.SSHPrivateKey)
	rec.SSHKeyPass, _ = secret.EncryptPtr(p.SSHKeyPass)
	rec.SudoPassword, _ = secret.EncryptPtr(p.SudoPassword)
	if err := db.ShellConnection.Create(rec); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to create shell connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Created shell connection", fmt.Sprintf("%s (%s)", rec.Name, rec.Mode), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"id": rec.ID})
}

func updateShellConnection(c *gin.Context) {
	id := atoi64(c.Param("id"))
	var p shellConnPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	if err := validateShellConnPayload(&p, true); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, err.Error())
		return
	}
	rec, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Shell connection not found")
		return
	}
	if p.Name != nil {
		rec.Name = *p.Name
	}
	if p.Description != nil {
		rec.Description = p.Description
	}
	if p.AgentUUID != nil {
		rec.AgentUUID = p.AgentUUID
	}
	if p.Mode != nil {
		rec.Mode = *p.Mode
	}
	if p.RunAsUser != nil {
		rec.RunAsUser = p.RunAsUser
	}
	if p.Sudo != nil {
		sudoVal := *p.Sudo
		rec.Sudo = &sudoVal
	}
	if p.Host != nil {
		rec.Host = p.Host
	}
	if p.Port != nil {
		rec.Port = p.Port
	}
	if p.SSHUsername != nil {
		rec.SSHUsername = p.SSHUsername
	}
	if p.AuthMethod != nil {
		rec.AuthMethod = p.AuthMethod
	}
	if p.AutoCheckEnabled != nil {
		if *p.AutoCheckEnabled {
			rec.AutoCheckEnabled = utilities.Ptr(int64(1))
		} else {
			rec.AutoCheckEnabled = utilities.Ptr(int64(0))
		}
	}
	if p.AutoCheckIntervalSeconds > 0 {
		rec.AutoCheckIntervalSeconds = p.AutoCheckIntervalSeconds
	}

	rec.AlertEmails = p.AlertEmails
	rec.AlertPhones = p.AlertPhones
	rec.NotifyOnFailure = p.NotifyOnFailure

	if p.Enabled != nil {
		if *p.Enabled && !rec.Enabled {
		}
		rec.Enabled = *p.Enabled
	}
	if p.Suspended != nil {
		rec.Suspended = *p.Suspended
	}
	if p.SSHPassword != nil && *p.SSHPassword != "<redacted>" {
		rec.SSHPassword, _ = secret.EncryptPtr(p.SSHPassword)
	}
	if p.SSHPrivateKey != nil && *p.SSHPrivateKey != "<redacted>" {
		rec.SSHPrivateKey, _ = secret.EncryptPtr(p.SSHPrivateKey)
	}
	if p.SSHKeyPass != nil && *p.SSHKeyPass != "<redacted>" {
		rec.SSHKeyPass, _ = secret.EncryptPtr(p.SSHKeyPass)
	}
	if p.SudoPassword != nil && *p.SudoPassword != "<redacted>" {
		rec.SudoPassword, _ = secret.EncryptPtr(p.SudoPassword)
	}
	rec.UpdatedAt = time.Now().UTC()
	if err := db.ShellConnection.Save(rec); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to update shell connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Updated shell connection", rec.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func deleteShellConnection(c *gin.Context) {
	id := atoi64(c.Param("id"))
	item, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Shell connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load connection", err.Error())
		return
	}
	usageCount, err := db.Job.Where(db.Job.ShellConnectionID.Eq(id), db.Job.Enabled.Is(true)).Count()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to check usage", err.Error())
		return
	}
	if usageCount > 0 {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Cannot delete connection: it is referenced by enabled jobs")
		return
	}

	if _, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete shell connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Deleted shell connection", item.Name, c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccessNoContent(c)
}

func clearShellConnectionSecret(c *gin.Context) {
	id := atoi64(c.Param("id"))
	var p struct {
		Field string `json:"field"`
	}
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	rec, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "Shell connection not found")
		return
	}

	switch p.Field {
	case "ssh_password":
		rec.SSHPassword = nil
	case "ssh_private_key":
		rec.SSHPrivateKey = nil
	case "ssh_key_pass":
		rec.SSHKeyPass = nil
	case "sudo_password":
		rec.SudoPassword = nil
	default:
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid field to clear")
		return
	}

	rec.UpdatedAt = time.Now().UTC()
	if err := db.ShellConnection.Save(rec); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to clear secret", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func duplicateShellConnection(c *gin.Context) {
	id := c.Param("id")
	item, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(atoi64(id))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Shell connection not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load shell connection", err.Error())
		return
	}

	newItem := *item
	newItem.ID = nil
	newItem.Name = "Copy Of " + item.Name
	newItem.CreatedAt = time.Now()
	newItem.UpdatedAt = time.Now()
	newItem.LastStatus = nil
	newItem.LastError = nil
	newItem.LastCheckedAt = nil

	if err := db.ShellConnection.Create(&newItem); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to duplicate shell connection", err.Error())
		return
	}
	user := userFromGinContext(c)
	_ = activitypkg.RecordUserActivity(user.ID, "Duplicated shell connection", fmt.Sprintf("%s -> %s", item.Name, newItem.Name), c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, redactShellConn(&newItem))
}
