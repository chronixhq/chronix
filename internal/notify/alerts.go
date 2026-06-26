package notify

import (
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/notifier"
	"chronix/pkg/typeutil"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/dan-sherwin/go-utilities"
)

// DispatchAlerts evaluates alert rules against a notification and sends emails/SMS as configured.
func DispatchAlerts(n *models.Notification) {
	cxs := cxsettingspkg.GetCxSettings()

	emailReady := utilities.PtrVal(cxs.SMTPHost) != ""
	smsReady := (utilities.PtrVal(cxs.SmsProvider) == "twilio" && (utilities.PtrVal(cxs.TwilioAccountSid) != "" || utilities.PtrVal(cxs.TwilioUsername) != ""))

	if !emailReady && !smsReady {
		return
	}

	emails := []string{}
	phones := []string{}

	// 1. Gather global system alert recipients for Error/Warning
	if n.Severity == string(SeverityError) || n.Severity == string(SeverityWarning) {
		if utilities.PtrVal(cxs.SystemAlertEmails) != "" {
			emails = append(emails, strings.Split(utilities.PtrVal(cxs.SystemAlertEmails), ",")...)
		}
		if utilities.PtrVal(cxs.SystemAlertPhones) != "" {
			phones = append(phones, strings.Split(utilities.PtrVal(cxs.SystemAlertPhones), ",")...)
		}
	}

	// 2. Gather category-specific recipients
	switch NotificationCategory(n.Category) {
	case CategoryJob:
		if n.Data != nil {
			var jobID int64
			if idVal, ok := (*n.Data)["job_id"]; ok {
				jobID = typeutil.AsInt64(idVal)
			}
			if jobID > 0 {
				if job, err := db.Job.Where(db.Job.ID.Eq(jobID)).First(); err == nil && job != nil {
					// Ensure job_name is present in data for subject/header construction
					if _, ok := (*n.Data)["job_name"]; !ok {
						(*n.Data)["job_name"] = job.Name
					}

					shouldNotify := false
					if n.Severity == string(SeveritySuccess) && utilities.PtrVal(job.NotifyOnSuccess) {
						shouldNotify = true
					} else if n.Severity == string(SeverityError) && utilities.PtrVal(job.NotifyOnError) {
						shouldNotify = true
					}

					if shouldNotify {
						added := false
						if utilities.PtrVal(job.AlertEmails) != "" {
							emails = append(emails, strings.Split(utilities.PtrVal(job.AlertEmails), ",")...)
							added = true
						}
						if utilities.PtrVal(job.AlertPhones) != "" {
							phones = append(phones, strings.Split(utilities.PtrVal(job.AlertPhones), ",")...)
							added = true
						}

						// If job has no specific overrides, fallback to connection's alerts
						if !added {
							connID := utilities.PtrVal(job.ConnectionID)
							kind := job.TargetKind
							switch kind {
							case "shell":
								connID = utilities.PtrVal(job.ShellConnectionID)
							case "webtask":
								connID = utilities.PtrVal(job.WebtaskConnectionID)
							}

							if connID > 0 {
								var cEmails *string
								var cPhones *string
								switch kind {
								case "database", "":
									if c, err := db.DbConnection.Where(db.DbConnection.ID.Eq(connID)).First(); err == nil && c != nil {
										cEmails = c.AlertEmails
										cPhones = c.AlertPhones
									}
								case "shell":
									if c, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(connID)).First(); err == nil && c != nil {
										cEmails = c.AlertEmails
										cPhones = c.AlertPhones
									}
								case "webtask":
									if c, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(connID)).First(); err == nil && c != nil {
										cEmails = c.AlertEmails
										cPhones = c.AlertPhones
									}
								}
								if utilities.PtrVal(cEmails) != "" {
									emails = append(emails, strings.Split(utilities.PtrVal(cEmails), ",")...)
								}
								if utilities.PtrVal(cPhones) != "" {
									phones = append(phones, strings.Split(utilities.PtrVal(cPhones), ",")...)
								}
							}
						}
					}
				}
			}
		}
	case CategoryConnection:
		if n.Data != nil {
			var connID int64
			if idVal, ok := (*n.Data)["connection_id"]; ok {
				connID = typeutil.AsInt64(idVal)
			}
			if connID > 0 {
				kind, _ := (*n.Data)["kind"].(string)

				var alertEmails *string
				var alertPhones *string
				var notifyOnFailure *bool
				var connName string

				if kind == "database" || kind == "" {
					if c, err := db.DbConnection.Where(db.DbConnection.ID.Eq(connID)).First(); err == nil && c != nil {
						alertEmails = c.AlertEmails
						alertPhones = c.AlertPhones
						notifyOnFailure = c.NotifyOnFailure
						connName = c.Name
					}
				}
				if (alertEmails == nil) && (kind == "shell" || kind == "") {
					if c, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(connID)).First(); err == nil && c != nil {
						alertEmails = c.AlertEmails
						alertPhones = c.AlertPhones
						notifyOnFailure = c.NotifyOnFailure
						connName = c.Name
					}
				}
				if (alertEmails == nil) && (kind == "webtask" || kind == "") {
					if c, err := db.WebtaskConnection.Where(db.WebtaskConnection.ID.Eq(connID)).First(); err == nil && c != nil {
						alertEmails = c.AlertEmails
						alertPhones = c.AlertPhones
						notifyOnFailure = c.NotifyOnFailure
						connName = c.Name
					}
				}

				if connName != "" {
					if _, ok := (*n.Data)["connection_name"]; !ok {
						(*n.Data)["connection_name"] = connName
					}
				}

				// For connections, we alert on Error if NotifyOnFailure is true (default true)
				if n.Severity == string(SeverityError) && (notifyOnFailure == nil || *notifyOnFailure) {
					if utilities.PtrVal(alertEmails) != "" {
						emails = append(emails, strings.Split(utilities.PtrVal(alertEmails), ",")...)
					}
					if utilities.PtrVal(alertPhones) != "" {
						phones = append(phones, strings.Split(utilities.PtrVal(alertPhones), ",")...)
					}
				}
			}
		}
	}

	// 3. Deduplicate and clean
	emails = cleanAndDedupe(emails)
	phones = cleanAndDedupe(phones)

	// 4. Send Emails
	if emailReady && len(emails) > 0 {
		cfg := notifier.EmailConfig{
			Host:      utilities.PtrVal(cxs.SMTPHost),
			Port:      int(utilities.PtrVal(cxs.SMTPPort)),
			Login:     utilities.PtrVal(cxs.SMTPLogin),
			Password:  utilities.PtrVal(cxs.SMTPPassword),
			FromName:  utilities.PtrVal(cxs.SMTPFromName),
			FromEmail: utilities.PtrVal(cxs.SMTPFromEmail),
			Secure:    utilities.PtrVal(cxs.SMTPSecure),
		}
		subject := fmt.Sprintf("[%s] %s", strings.ToUpper(n.Severity), n.Subject)
		if n.Category == string(CategoryJob) && n.Data != nil {
			jobName, _ := (*n.Data)["job_name"].(string)
			status, _ := (*n.Data)["status"].(string)
			if status == "" {
				status = n.Severity
			}
			if jobName != "" && status != "" {
				subject = fmt.Sprintf("Chronix alert: %s: %s", jobName, strings.ToUpper(status))
			}
		} else if n.Category == string(CategoryConnection) && n.Data != nil {
			connName, _ := (*n.Data)["connection_name"].(string)
			status, _ := (*n.Data)["status"].(string) // try status from health check
			if status == "" {
				status = n.Severity
			}
			if connName != "" && status != "" {
				subject = fmt.Sprintf("Chronix alert: %s: %s", connName, strings.ToUpper(status))
			}
		}
		body := n.Subject
		if n.Origin != nil {
			body += "\nOrigin: " + *n.Origin
		}
		if n.Data != nil {
			body += "\n\nDetails:\n"
			// Sort keys for consistent output
			keys := make([]string, 0, len(*n.Data))
			for k := range *n.Data {
				if k != "output" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)

			for _, k := range keys {
				v := (*n.Data)[k]
				label := k
				switch k {
				case "job_id":
					label = "Job ID"
				case "job_name":
					label = "Job Name"
				case "run_id":
					label = "Run ID"
				case "status":
					label = "Status"
				case "message":
					label = "Message"
				case "finished_at":
					label = "Finished At"
				}
				body += fmt.Sprintf("- %s: %v\n", label, v)
			}
			if outputVal, ok := (*n.Data)["output"]; ok {
				if output, ok := asMap(outputVal); ok {
					body += formatOutputForEmail(output)
				}
			}
		}

		htmlBody := generateEmailHTML(n)
		go func() {
			if err := notifier.SendEmail(cfg, emails, subject, body, htmlBody); err != nil {
				slog.Error("dispatch email alerts", "error", err)
			}
		}()
	}

	// 5. Send SMS
	if smsReady && len(phones) > 0 {
		accountSid := utilities.PtrVal(cxs.TwilioAccountSid)
		if accountSid == "" {
			accountSid = utilities.PtrVal(cxs.TwilioUsername)
		}
		authToken := utilities.PtrVal(cxs.TwilioAPISecret)
		if authToken == "" {
			authToken = utilities.PtrVal(cxs.TwilioPassword)
		}

		cfg := notifier.SMSConfig{
			AccountSid: accountSid,
			AuthToken:  authToken,
			FromNumber: utilities.PtrVal(cxs.SmsFromNumber),
		}
		message := fmt.Sprintf("[%s] %s", strings.ToUpper(n.Severity), n.Subject)
		if n.Category == string(CategoryJob) && n.Data != nil {
			jobName, _ := (*n.Data)["job_name"].(string)
			status, _ := (*n.Data)["status"].(string)
			if status == "" {
				status = n.Severity
			}
			if jobName != "" && status != "" {
				message = fmt.Sprintf("Chronix alert: %s: %s", jobName, strings.ToUpper(status))
			}
		} else if n.Category == string(CategoryConnection) && n.Data != nil {
			connName, _ := (*n.Data)["connection_name"].(string)
			status, _ := (*n.Data)["status"].(string)
			if status == "" {
				status = n.Severity
			}
			if connName != "" && status != "" {
				message = fmt.Sprintf("Chronix alert: %s: %s", connName, strings.ToUpper(status))
			}
		}
		if n.Data != nil {
			if outputVal, ok := (*n.Data)["output"]; ok {
				if output, ok := asMap(outputVal); ok {
					message += "\n" + formatOutputForSMS(output)
				}
			}
		}

		go func() {
			if err := notifier.SendSMS(cfg, phones, message); err != nil {
				slog.Error("dispatch sms alerts", "error", err)
			}
		}()
	}

	// 6. Send Webhooks
	go func() {
		webhooks, err := db.Webhook.Where(db.Webhook.Enabled.Is(true)).Find()
		if err != nil {
			slog.Error("fetch webhooks for dispatch", "error", err)
			return
		}

		for _, wh := range webhooks {
			subscribedEvents := strings.Split(wh.Events, ",")
			isSubscribed := false
			for _, e := range subscribedEvents {
				if strings.TrimSpace(e) == n.Category {
					isSubscribed = true
					break
				}
			}

			if isSubscribed {
				cfg := notifier.WebhookConfig{
					URL:    wh.URL,
					Secret: utilities.PtrVal(wh.Secret),
				}
				if err := notifier.SendWebhook(cfg, n.Category, n); err != nil {
					slog.Error("dispatch webhook alert", "webhook", wh.Name, "error", err)
				}
			}
		}
	}()
}

func cleanAndDedupe(in []string) []string {
	m := make(map[string]bool)
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s != "" && !m[s] {
			m[s] = true
			out = append(out, s)
		}
	}
	return out
}
