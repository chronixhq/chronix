package webtaskconnectionsapi

import (
	"errors"
	"fmt"
	"strings"

	"chronix/internal/db/models"

	"github.com/gin-gonic/gin"
)

type webtaskConnPayload struct {
	ID               any            `json:"id"`
	Name             *string        `json:"name"`
	Description      *string        `json:"description"`
	BaseURL          *string        `json:"baseUrl"`
	AuthType         *string        `json:"authType"`
	AuthConfig       map[string]any `json:"authConfig"`
	AgentUUID        *string        `json:"agentUuid"`
	AutoCheckEnabled any            `json:"autoCheckEnabled"`
	AutoCheckSeconds any            `json:"autoCheckSeconds"`
	AlertEmails      *string        `json:"alertEmails"`
	AlertPhones      *string        `json:"alertPhones"`
	NotifyOnFailure  *bool          `json:"notifyOnFailure"`
	Enabled          *bool          `json:"enabled"`
	Suspended        *bool          `json:"suspended"`
	CreatedAt        any            `json:"createdAt"`
	UpdatedAt        any            `json:"updatedAt"`
	LastStatus       any            `json:"lastStatus"`
	LastError        any            `json:"lastError"`
	LastCheckedAt    any            `json:"lastCheckedAt"`
}

func validateWebtaskConnPayload(p *webtaskConnPayload) error {
	if p.Name == nil || strings.TrimSpace(*p.Name) == "" {
		return errors.New("name is required")
	}
	if p.AuthType == nil || strings.TrimSpace(*p.AuthType) == "" {
		return errors.New("authType is required")
	}
	return nil
}

func MapWebtaskConnection(it *models.WebtaskConnection) gin.H {
	return gin.H{
		"id":               it.ID,
		"name":             it.Name,
		"description":      it.Description,
		"baseUrl":          it.BaseURL,
		"authType":         it.AuthType,
		"authConfig":       it.AuthConfig,
		"agentUuid":        it.AgentUUID,
		"autoCheckEnabled": it.AutoCheckEnabled != nil && *it.AutoCheckEnabled != 0,
		"autoCheckSeconds": it.AutoCheckIntervalSeconds,
		"alertEmails":      it.AlertEmails,
		"alertPhones":      it.AlertPhones,
		"notifyOnFailure":  it.NotifyOnFailure,
		"enabled":          it.Enabled,
		"suspended":        it.Suspended,
		"createdAt":        it.CreatedAt,
		"updatedAt":        it.UpdatedAt,
		"lastStatus":       it.LastStatus,
		"lastError":        it.LastError,
		"lastCheckedAt":    it.LastCheckedAt,
	}
}

func AnyToI64(v any, def int64) int64 {
	if v == nil {
		return def
	}
	switch val := v.(type) {
	case bool:
		if val {
			return 1
		}
		return 0
	case float64:
		return int64(val)
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		var out int64
		if _, err := fmt.Sscanf(val, "%d", &out); err != nil {
			return def
		}
		return out
	}
	return def
}
