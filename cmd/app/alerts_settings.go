package app

import (
	"fmt"
	"strconv"
	"time"

	app_settings "github.com/dan-sherwin/go-app-settings"

	notifypkg "chronix/internal/notify"
)

// Register alert/notification related tuning knobs. These only expose thresholds and controls
// for operational signals. Audit-grade alerts (security/audit/user events) are ALWAYS ON and
// intentionally do not have disable toggles here.
func init() {
	// Job scheduling/health
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid job miss grace: %v", err)
			}
			notifypkg.SetJobMissGrace(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetJobMissGrace().String() },
		Name:        "alert_job_miss_grace",
		Description: "Duration after scheduled time before a job is considered 'missed'",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return fmt.Errorf("invalid consecutive failures threshold")
			}
			notifypkg.SetJobConsecutiveFailuresError(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetJobConsecutiveFailuresError()) },
		Name:        "alert_job_consecutive_failures_error",
		Description: "Raise error after N consecutive job failures",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid job runtime warning: %v", err)
			}
			notifypkg.SetJobRuntimeWarning(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetJobRuntimeWarning().String() },
		Name:        "alert_job_runtime_warning",
		Description: "Warn if a job runtime exceeds this duration",
	})

	// Connection flapping
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid flap window: %v", err)
			}
			notifypkg.SetConnectionFlapWindow(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetConnectionFlapWindow().String() },
		Name:        "alert_connection_flap_window",
		Description: "Sliding window to detect flapping for connections",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 2 {
				return fmt.Errorf("invalid flap change count")
			}
			notifypkg.SetConnectionFlapChangeCount(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetConnectionFlapChangeCount()) },
		Name:        "alert_connection_flap_change_count",
		Description: "Number of state changes within window to consider a connection flapping",
	})

	// Agent lifecycle
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid agent heartbeat grace: %v", err)
			}
			notifypkg.SetAgentHeartbeatGrace(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetAgentHeartbeatGrace().String() },
		Name:        "alert_agent_heartbeat_grace",
		Description: "Grace period after last agent heartbeat before warnings/errors",
	})

	// System/storage
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 99 {
				return fmt.Errorf("invalid disk free warn percent")
			}
			notifypkg.SetDiskFreeWarnPercent(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetDiskFreeWarnPercent()) },
		Name:        "alert_disk_free_warn_percent",
		Description: "Warn when disk free percentage drops below this value",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 || n > 99 {
				return fmt.Errorf("invalid disk free error percent")
			}
			notifypkg.SetDiskFreeErrorPercent(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetDiskFreeErrorPercent()) },
		Name:        "alert_disk_free_error_percent",
		Description: "Error when disk free percentage drops below this value",
	})

	// Certificate warning windows
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid cert warn duration: %v", err)
			}
			notifypkg.SetCertWarn(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetCertWarn().String() },
		Name:        "alert_cert_warn",
		Description: "Warn duration before certificate expiration",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			d, err := time.ParseDuration(s)
			if err != nil {
				return fmt.Errorf("invalid cert error duration: %v", err)
			}
			notifypkg.SetCertError(d)
			return nil
		},
		GetFunc:     func() string { return notifypkg.GetCertError().String() },
		Name:        "alert_cert_error",
		Description: "Error duration before certificate expiration",
	})

	// Security/auth signals thresholds
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return fmt.Errorf("invalid login failed attempts warn threshold")
			}
			notifypkg.SetLoginFailedAttemptsWarn(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetLoginFailedAttemptsWarn()) },
		Name:        "alert_login_failed_attempts_warn",
		Description: "Warn after this many failed login attempts within a window",
	})

	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return fmt.Errorf("invalid login failed attempts error threshold")
			}
			notifypkg.SetLoginFailedAttemptsError(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetLoginFailedAttemptsError()) },
		Name:        "alert_login_failed_attempts_error",
		Description: "Error after this many failed login attempts within a window",
	})

	// Flood control
	app_settings.RegisterSetting(&app_settings.Setting{
		SetFunc: func(s string) error {
			n, err := strconv.Atoi(s)
			if err != nil || n < 1 {
				return fmt.Errorf("invalid per-source per-minute rate limit")
			}
			notifypkg.SetRateLimitPerSourcePerMin(n)
			return nil
		},
		GetFunc:     func() string { return strconv.Itoa(notifypkg.GetRateLimitPerSourcePerMin()) },
		Name:        "alert_rate_limit_per_source_per_min",
		Description: "Soft cap for notifications per source per minute (flood control)",
	})
}
