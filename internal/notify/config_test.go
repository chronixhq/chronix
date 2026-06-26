package notify

import (
	"testing"
	"time"
)

func TestAlertsConfig_GettersSetters(t *testing.T) {
	// Save originals to restore
	origMiss := GetJobMissGrace()
	origErrThresh := GetJobConsecutiveFailuresError()
	origRuntimeWarn := GetJobRuntimeWarning()

	origFlapWin := GetConnectionFlapWindow()
	origFlapCnt := GetConnectionFlapChangeCount()
	origHB := GetAgentHeartbeatGrace()
	origWarn := GetDiskFreeWarnPercent()
	origErr := GetDiskFreeErrorPercent()
	origCertWarn := GetCertWarn()
	origCertErr := GetCertError()
	origLoginWarn := GetLoginFailedAttemptsWarn()
	origLoginErr := GetLoginFailedAttemptsError()
	origRate := GetRateLimitPerSourcePerMin()

	defer func() {
		SetJobMissGrace(origMiss)
		SetJobConsecutiveFailuresError(origErrThresh)
		SetJobRuntimeWarning(origRuntimeWarn)
		SetConnectionFlapWindow(origFlapWin)
		SetConnectionFlapChangeCount(origFlapCnt)
		SetAgentHeartbeatGrace(origHB)
		SetDiskFreeWarnPercent(origWarn)
		SetDiskFreeErrorPercent(origErr)
		SetCertWarn(origCertWarn)
		SetCertError(origCertErr)
		SetLoginFailedAttemptsWarn(origLoginWarn)
		SetLoginFailedAttemptsError(origLoginErr)
		SetRateLimitPerSourcePerMin(origRate)
	}()

	SetJobMissGrace(321 * time.Second)
	if got := GetJobMissGrace(); got != 321*time.Second {
		t.Fatalf("JobMissGrace got=%v want=321s", got)
	}
	SetJobConsecutiveFailuresError(9)
	if got := GetJobConsecutiveFailuresError(); got != 9 {
		t.Fatalf("JobConsecutiveFailuresError got=%d want=9", got)
	}
	SetJobRuntimeWarning(777 * time.Second)
	if got := GetJobRuntimeWarning(); got != 777*time.Second {
		t.Fatalf("JobRuntimeWarning got=%v want=777s", got)
	}

	SetConnectionFlapWindow(123 * time.Second)
	SetConnectionFlapChangeCount(7)
	if GetConnectionFlapWindow() != 123*time.Second || GetConnectionFlapChangeCount() != 7 {
		t.Fatalf("flap config not set correctly")
	}

	SetAgentHeartbeatGrace(42 * time.Second)
	if GetAgentHeartbeatGrace() != 42*time.Second {
		t.Fatalf("AgentHeartbeatGrace not set")
	}

	SetDiskFreeWarnPercent(11)
	SetDiskFreeErrorPercent(3)
	if GetDiskFreeWarnPercent() != 11 || GetDiskFreeErrorPercent() != 3 {
		t.Fatalf("disk free thresholds not set")
	}

	SetCertWarn(20 * 24 * time.Hour)
	SetCertError(2 * 24 * time.Hour)
	if GetCertWarn() != 20*24*time.Hour || GetCertError() != 2*24*time.Hour {
		t.Fatalf("cert thresholds not set")
	}

	SetLoginFailedAttemptsWarn(4)
	SetLoginFailedAttemptsError(8)
	if GetLoginFailedAttemptsWarn() != 4 || GetLoginFailedAttemptsError() != 8 {
		t.Fatalf("login thresholds not set")
	}

	SetRateLimitPerSourcePerMin(55)
	if GetRateLimitPerSourcePerMin() != 55 {
		t.Fatalf("rate limit not set")
	}
}
