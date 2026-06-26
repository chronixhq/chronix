package notifier

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	mail "github.com/xhit/go-simple-mail/v2"
)

type EmailConfig struct {
	Host      string
	Port      int
	Login     string
	Password  string
	FromName  string
	FromEmail string
	Secure    string // "none", "ssl", "starttls"
}

func SendEmail(cfg EmailConfig, to []string, subject, plainBody, htmlBody string) error {
	if cfg.Host == "" {
		return fmt.Errorf("SMTP host not provided")
	}

	server := mail.NewSMTPClient()
	server.Host = cfg.Host
	server.Port = cfg.Port
	server.Username = cfg.Login
	server.Password = cfg.Password
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	switch cfg.Secure {
	case "ssl":
		server.Encryption = mail.EncryptionSSLTLS
	case "starttls":
		server.Encryption = mail.EncryptionSTARTTLS
	default:
		server.Encryption = mail.EncryptionNone
	}

	// Many corporate or internal SMTP servers use self-signed certificates.
	// We allow insecure skip verify to ensure connectivity in these environments.
	server.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	smtpClient, err := server.Connect()
	if err != nil {
		return fmt.Errorf("connect to smtp server: %w", err)
	}
	defer func() { _ = smtpClient.Close() }()

	email := mail.NewMSG()

	from := cfg.FromEmail
	if cfg.FromName != "" {
		from = fmt.Sprintf("%s <%s>", cfg.FromName, cfg.FromEmail)
	}

	email.SetFrom(from).
		AddTo(to...).
		SetSubject(subject)

	if plainBody != "" && htmlBody != "" {
		email.SetBody(mail.TextPlain, plainBody)
		email.AddAlternative(mail.TextHTML, htmlBody)
	} else if htmlBody != "" {
		email.SetBody(mail.TextHTML, htmlBody)
	} else {
		email.SetBody(mail.TextPlain, plainBody)
	}

	// Add Message-ID header to satisfy Gmail and other strict providers.
	domain := "chronix"
	if parts := strings.Split(cfg.FromEmail, "@"); len(parts) > 1 {
		domain = parts[1]
	}
	msgID := fmt.Sprintf("<%s@%s>", uuid.New().String(), domain)
	email.AddHeader("Message-ID", msgID)

	if email.Error != nil {
		return fmt.Errorf("create email message: %w", email.Error)
	}

	err = email.Send(smtpClient)
	if err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}
