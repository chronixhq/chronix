package cxsettings

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/secret"
	"chronix/internal/updater"

	"github.com/dan-sherwin/go-utilities"
)

var (
	CxSettings models.CxSetting
)

// LoadCxSettings fetches the system configuration from the database and decrypts sensitive fields.
func LoadCxSettings() {
	if cxs, _ := db.CxSetting.Where(db.CxSetting.ID.Eq(0)).First(); cxs != nil {
		CxSettings = *cxs
		// Decrypt sensitive fields
		CxSettings.SMTPPassword, _ = secret.DecryptPtr(CxSettings.SMTPPassword)
		CxSettings.TwilioPassword, _ = secret.DecryptPtr(CxSettings.TwilioPassword)
		CxSettings.TwilioAPISecret, _ = secret.DecryptPtr(CxSettings.TwilioAPISecret)
		CxSettings.HTTPSKeyPem, _ = secret.DecryptPtr(CxSettings.HTTPSKeyPem)
		SyncUpdaterSettings()
	}
}

func GetCxSettings() models.CxSetting {
	return CxSettings
}

func SetCxSettings() error {
	CxSettings.ID = utilities.Ptr(int64(0))

	// Create a copy for encryption
	toSave := CxSettings
	toSave.SMTPPassword, _ = secret.EncryptPtr(toSave.SMTPPassword)
	toSave.TwilioPassword, _ = secret.EncryptPtr(toSave.TwilioPassword)
	toSave.TwilioAPISecret, _ = secret.EncryptPtr(toSave.TwilioAPISecret)
	toSave.HTTPSKeyPem, _ = secret.EncryptPtr(toSave.HTTPSKeyPem)

	err := db.CxSetting.Save(&toSave)
	if err != nil {
		return err
	}
	return nil
}

func SyncUpdaterSettings() {
	if CxSettings.UpdaterEnabled != nil {
		updater.Enabled = *CxSettings.UpdaterEnabled
	}
	if mode := utilities.PtrVal(CxSettings.UpdaterMode); mode != "" {
		updater.Mode = mode
	}
	updater.WindowStart = utilities.PtrVal(CxSettings.UpdaterWindowStart)

	if CxSettings.UpdaterAgentEnabled != nil {
		updater.AgentEnabled = *CxSettings.UpdaterAgentEnabled
	}
	if mode := utilities.PtrVal(CxSettings.UpdaterAgentMode); mode != "" {
		updater.AgentMode = mode
	}
	updater.AgentWindowStart = utilities.PtrVal(CxSettings.UpdaterAgentWindowStart)
}
