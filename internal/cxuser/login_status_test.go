package cxuser

import (
	"chronix/internal/db"
	"os"
	"testing"
)

func TestLogin_SuspendedUser(t *testing.T) {
	// Setup temporary DB
	tempDB := "repro_login.db"
	if _, err := db.EnsureSQLiteFile(tempDB); err != nil {
		t.Fatalf("Failed to ensure sqlite file: %v", err)
	}
	db.DbInit(tempDB)
	defer func() { _ = os.Remove(tempDB) }()

	password := "testpassword123"
	user := &CxUser{}
	user.Email = "suspended@example.com"
	user.Name = "Suspended User"
	user.Password = &password
	if err := user.Save(); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Now mark as suspended and ensure it's enabled
	if _, err := db.CxUser.Where(db.CxUser.ID.Eq(user.ID)).UpdateSimple(
		db.CxUser.Suspended.Value(true),
		db.CxUser.Enabled.Value(true),
	); err != nil {
		t.Fatalf("Failed to update user flags: %v", err)
	}

	// Try to login
	loggedUser, err := Login("suspended@example.com", password)
	if err != nil {
		// If it fails, maybe it's already fixed or failing for other reasons?
		// But currently it should SUCCEED (return nil error) which is the BUG.
		t.Logf("Login failed as expected if fixed: %v", err)
	} else {
		if loggedUser != nil {
			t.Errorf("Expected login to fail for suspended user, but it succeeded")
		}
	}
}

func TestLogin_DisabledUser(t *testing.T) {
	// Setup temporary DB
	tempDB := "repro_login_disabled.db"
	if _, err := db.EnsureSQLiteFile(tempDB); err != nil {
		t.Fatalf("Failed to ensure sqlite file: %v", err)
	}
	db.DbInit(tempDB)
	defer func() { _ = os.Remove(tempDB) }()

	// Create a disabled user
	password := "testpassword123"
	user := &CxUser{}
	user.Email = "disabled@example.com"
	user.Name = "Disabled User"
	user.Password = &password
	if err := user.Save(); err != nil {
		t.Fatalf("Failed to save user: %v", err)
	}

	// Now mark as disabled
	if _, err := db.CxUser.Where(db.CxUser.ID.Eq(user.ID)).Update(db.CxUser.Enabled, false); err != nil {
		t.Fatalf("Failed to update user flags: %v", err)
	}

	// Try to login
	loggedUser, err := Login("disabled@example.com", password)
	if err == nil && loggedUser != nil {
		t.Errorf("Expected login to fail for disabled user, but it succeeded")
	} else {
		t.Logf("Login failed as expected: %v", err)
	}
}
