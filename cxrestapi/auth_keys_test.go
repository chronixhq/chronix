package cxrestapi

import (
	cxuserpkg "chronix/internal/cxuser"
	"chronix/internal/db"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestSyncAuthKeys_SuspendedAndDisabled(t *testing.T) {
	// Setup temporary DB
	tempDB := "repro_sync_keys.db"
	if _, err := db.EnsureSQLiteFile(tempDB); err != nil {
		t.Fatalf("Failed to ensure sqlite file: %v", err)
	}
	db.DbInit(tempDB)
	defer func() { _ = os.Remove(tempDB) }()

	// Clear existing auth keys in map
	authKeys = make(map[string]struct{})

	// Create 3 users: active, suspended, disabled
	password := "testpassword123"

	createTestUser := func(email, name string, enabled, suspended bool) int64 {
		user := &cxuserpkg.CxUser{}
		user.Email = email
		user.Name = name
		user.Password = &password
		if err := user.Save(); err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}
		if _, err := db.CxUser.Where(db.CxUser.ID.Eq(user.ID)).UpdateSimple(
			db.CxUser.Enabled.Value(enabled),
			db.CxUser.Suspended.Value(suspended),
		); err != nil {
			t.Fatalf("Failed to update user flags: %v", err)
		}
		return user.ID
	}

	idActive := createTestUser("active@example.com", "Active", true, false)
	idSuspended := createTestUser("suspended@example.com", "Suspended", true, true)
	idDisabled := createTestUser("disabled@example.com", "Disabled", false, false)

	// Add keys to authKeys map
	keyActive := fmt.Sprintf("%d:0:%d", idActive, time.Now().UnixMilli())
	keySuspended := fmt.Sprintf("%d:0:%d", idSuspended, time.Now().UnixMilli())
	keyDisabled := fmt.Sprintf("%d:0:%d", idDisabled, time.Now().UnixMilli())

	authKeys[keyActive] = struct{}{}
	authKeys[keySuspended] = struct{}{}
	authKeys[keyDisabled] = struct{}{}

	// Run SyncAuthKeys
	if err := SyncAuthKeys(); err != nil {
		t.Fatalf("SyncAuthKeys failed: %v", err)
	}

	// Verify
	if _, ok := authKeys[keyActive]; !ok {
		t.Errorf("Expected active user key to remain, but it was deleted")
	}
	if _, ok := authKeys[keySuspended]; ok {
		t.Errorf("Expected suspended user key to be deleted, but it remains")
	}
	if _, ok := authKeys[keyDisabled]; ok {
		t.Errorf("Expected disabled user key to be deleted, but it remains")
	}
}
