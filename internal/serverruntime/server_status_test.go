package serverruntime

import (
	cxsettingspkg "chronix/internal/cxsettings"
	"testing"
)

func TestUpdateServerStatus_Suspended(t *testing.T) {
	// Force suspend lock to short-circuit before DB access
	suspendLock = true
	defer func() { suspendLock = false }()
	CurrentServerStatus = StatusUnknown
	UpdateServerStatus()
	if CurrentServerStatus != StatusSuspended {
		t.Fatalf("want %s, got %s", StatusSuspended, CurrentServerStatus)
	}
}

func TestUpdateServerStatus_Uninitialized_NoServerURL(t *testing.T) {
	// No suspend; empty ServerURL forces uninitialized, short-circuiting before DB
	suspendLock = false
	CurrentServerStatus = StatusUnknown
	cxsettingspkg.CxSettings.ServerURL = nil
	UpdateServerStatus()
	if CurrentServerStatus != StatusUninitialized {
		t.Fatalf("want %s, got %s", StatusUninitialized, CurrentServerStatus)
	}
}
