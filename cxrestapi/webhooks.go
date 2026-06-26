package cxrestapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/notifier"
	"errors"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func webhooksRouter(app *gin.Engine) {
	g := app.Group("/webhooks")
	{
		g.GET("", adminFunc(listWebhooks))
		g.POST("", adminFunc(createWebhook))
		g.GET("/:id", adminFunc(getWebhook))
		g.PUT("/:id", adminFunc(updateWebhook))
		g.DELETE("/:id", adminFunc(deleteWebhook))
		g.POST("/test", adminFunc(testWebhook))
	}
}

type webhookPayload struct {
	ID      *int64 `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Secret  string `json:"secret"`
	Events  string `json:"events"`
	Enabled bool   `json:"enabled"`
}

func listWebhooks(c *gin.Context) {
	rows, err := db.Webhook.Order(db.Webhook.Name.Asc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to list webhooks", err.Error())
		return
	}
	restresponse.RestSuccess(c, rows)
}

func getWebhook(c *gin.Context) {
	id := atoi64(c.Param("id"))
	wh, err := db.Webhook.Where(db.Webhook.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Webhook not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load webhook", err.Error())
		return
	}
	restresponse.RestSuccess(c, wh)
}

func createWebhook(c *gin.Context) {
	var p webhookPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	if strings.TrimSpace(p.Name) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Name is required")
		return
	}
	if strings.TrimSpace(p.URL) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "URL is required")
		return
	}

	now := time.Now().UTC()
	wh := &models.Webhook{
		Name:      p.Name,
		URL:       p.URL,
		Secret:    utilities.Ptr(p.Secret),
		Events:    p.Events,
		Enabled:   &p.Enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := db.Webhook.Create(wh); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to create webhook", err.Error())
		return
	}

	restresponse.RestSuccess(c, wh)
}

func updateWebhook(c *gin.Context) {
	id := atoi64(c.Param("id"))
	var p webhookPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	wh, err := db.Webhook.Where(db.Webhook.ID.Eq(id)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			restresponse.RestErrorRespond(c, restresponse.NotFound, "Webhook not found")
			return
		}
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to load webhook", err.Error())
		return
	}

	if strings.TrimSpace(p.Name) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Name is required")
		return
	}
	if strings.TrimSpace(p.URL) == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "URL is required")
		return
	}

	wh.Name = p.Name
	wh.URL = p.URL
	wh.Secret = utilities.Ptr(p.Secret)
	wh.Events = p.Events
	wh.Enabled = &p.Enabled
	wh.UpdatedAt = time.Now().UTC()

	if err := db.Webhook.Save(wh); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to update webhook", err.Error())
		return
	}

	restresponse.RestSuccess(c, wh)
}

func deleteWebhook(c *gin.Context) {
	id := atoi64(c.Param("id"))
	if _, err := db.Webhook.Where(db.Webhook.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to delete webhook", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func testWebhook(c *gin.Context) {
	var p webhookPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	cfg := notifier.WebhookConfig{
		URL:    p.URL,
		Secret: p.Secret,
	}

	pingPayload := map[string]any{
		"ping":      true,
		"timestamp": time.Now().Unix(),
		"message":   "Chronix webhook test ping",
	}

	if err := notifier.SendWebhook(cfg, "ping", pingPayload); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Webhook test failed", err.Error())
		return
	}

	restresponse.RestSuccess(c, gin.H{"ok": true, "message": "Webhook test successful"})
}
