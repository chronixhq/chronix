package notify

import (
	"sync"
	"time"
)

// Alerts configuration holds thresholds and knobs that influence when alerts are emitted.
// Audit-grade alerts (security/user/audit events) are ALWAYS ON by policy and have no disable toggles here.
// All values are process-local runtime settings adjustable via app_settings (CLI/env/persisted) per guidelines.

type alertsCfg struct {
	mu sync.RWMutex

	JobMissGrace                time.Duration // how long after schedule time before we say "missed"
	JobConsecutiveFailuresError int           // error when this threshold of consecutive failures is reached
	JobRuntimeWarning           time.Duration // warn when a run exceeds this runtime

	ConnectionFlapWindow      time.Duration // sliding window to consider flapping
	ConnectionFlapChangeCount int           // number of state changes within window to call it flapping

	AgentHeartbeatGrace time.Duration // grace after last heartbeat before warning/error logic kicks in

	DiskFreeWarnPercent  int // warn when disk free % falls below this
	DiskFreeErrorPercent int // error when disk free % falls below this

	CertWarn  time.Duration // warn before cert expiration
	CertError time.Duration // error before cert expiration (typically 0 or 1)

	LoginFailedAttemptsWarn  int // failed logins triggering warning
	LoginFailedAttemptsError int // failed logins triggering error/escalation

	RateLimitPerSourcePerMin int // soft cap for notifications per source per minute (for flood control)
}

var _alerts = &alertsCfg{
	JobMissGrace:                120 * time.Second,
	JobConsecutiveFailuresError: 3,
	JobRuntimeWarning:           300 * time.Second,

	ConnectionFlapWindow:      300 * time.Second,
	ConnectionFlapChangeCount: 4,

	AgentHeartbeatGrace: 60 * time.Second,

	DiskFreeWarnPercent:  15,
	DiskFreeErrorPercent: 5,

	CertWarn:  30 * 24 * time.Hour,
	CertError: 1 * 24 * time.Hour,

	LoginFailedAttemptsWarn:  5,
	LoginFailedAttemptsError: 10,

	RateLimitPerSourcePerMin: 30,
}

// Getters
func GetJobMissGrace() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.JobMissGrace
}
func GetJobConsecutiveFailuresError() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.JobConsecutiveFailuresError
}
func GetJobRuntimeWarning() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.JobRuntimeWarning
}
func GetConnectionFlapWindow() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.ConnectionFlapWindow
}
func GetConnectionFlapChangeCount() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.ConnectionFlapChangeCount
}
func GetAgentHeartbeatGrace() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.AgentHeartbeatGrace
}
func GetDiskFreeWarnPercent() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.DiskFreeWarnPercent
}
func GetDiskFreeErrorPercent() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.DiskFreeErrorPercent
}
func GetCertWarn() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.CertWarn
}
func GetCertError() time.Duration {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.CertError
}
func GetLoginFailedAttemptsWarn() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.LoginFailedAttemptsWarn
}
func GetLoginFailedAttemptsError() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.LoginFailedAttemptsError
}
func GetRateLimitPerSourcePerMin() int {
	_alerts.mu.RLock()
	defer _alerts.mu.RUnlock()
	return _alerts.RateLimitPerSourcePerMin
}

// Setters
func SetJobMissGrace(d time.Duration) {
	_alerts.mu.Lock()
	_alerts.JobMissGrace = d
	_alerts.mu.Unlock()
}
func SetJobConsecutiveFailuresError(n int) {
	_alerts.mu.Lock()
	_alerts.JobConsecutiveFailuresError = n
	_alerts.mu.Unlock()
}
func SetJobRuntimeWarning(d time.Duration) {
	_alerts.mu.Lock()
	_alerts.JobRuntimeWarning = d
	_alerts.mu.Unlock()
}
func SetConnectionFlapWindow(d time.Duration) {
	_alerts.mu.Lock()
	_alerts.ConnectionFlapWindow = d
	_alerts.mu.Unlock()
}
func SetConnectionFlapChangeCount(n int) {
	_alerts.mu.Lock()
	_alerts.ConnectionFlapChangeCount = n
	_alerts.mu.Unlock()
}
func SetAgentHeartbeatGrace(d time.Duration) {
	_alerts.mu.Lock()
	_alerts.AgentHeartbeatGrace = d
	_alerts.mu.Unlock()
}
func SetDiskFreeWarnPercent(n int) {
	_alerts.mu.Lock()
	_alerts.DiskFreeWarnPercent = n
	_alerts.mu.Unlock()
}
func SetDiskFreeErrorPercent(n int) {
	_alerts.mu.Lock()
	_alerts.DiskFreeErrorPercent = n
	_alerts.mu.Unlock()
}
func SetCertWarn(d time.Duration)  { _alerts.mu.Lock(); _alerts.CertWarn = d; _alerts.mu.Unlock() }
func SetCertError(d time.Duration) { _alerts.mu.Lock(); _alerts.CertError = d; _alerts.mu.Unlock() }
func SetLoginFailedAttemptsWarn(n int) {
	_alerts.mu.Lock()
	_alerts.LoginFailedAttemptsWarn = n
	_alerts.mu.Unlock()
}
func SetLoginFailedAttemptsError(n int) {
	_alerts.mu.Lock()
	_alerts.LoginFailedAttemptsError = n
	_alerts.mu.Unlock()
}
func SetRateLimitPerSourcePerMin(n int) {
	_alerts.mu.Lock()
	_alerts.RateLimitPerSourcePerMin = n
	_alerts.mu.Unlock()
}
