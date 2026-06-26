package notifier

import (
	"fmt"
	"log/slog"

	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/api/v2010"
)

type SMSConfig struct {
	AccountSid string
	AuthToken  string // This is either the Auth Token or the API Secret
	FromNumber string
}

func SendSMS(cfg SMSConfig, to []string, message string) error {
	if cfg.AccountSid == "" {
		return fmt.Errorf("twilio account SID not provided")
	}
	if cfg.AuthToken == "" {
		return fmt.Errorf("twilio auth token / API secret not provided")
	}
	if cfg.FromNumber == "" {
		return fmt.Errorf("twilio 'from' number not provided")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: cfg.AccountSid,
		Password: cfg.AuthToken,
	})

	var lastErr error
	for _, phone := range to {
		params := &openapi.CreateMessageParams{}
		params.SetTo(phone)
		params.SetFrom(cfg.FromNumber)
		params.SetBody(message)

		resp, err := client.Api.CreateMessage(params)
		if err != nil {
			slog.Error("send SMS", "error", err, "to", phone)
			lastErr = err
			continue
		}

		if resp.Sid != nil {
			slog.Debug("SMS sent successfully", "sid", *resp.Sid, "to", phone)
		} else {
			slog.Warn("SMS sent but no SID returned", "to", phone)
		}
	}

	return lastErr
}
