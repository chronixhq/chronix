package db

import (
	"os"
	"strings"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestSyncSchema(t *testing.T) {
	// Create a temporary live database
	tmpLive, err := os.CreateTemp("", "live-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpLive.Name())
	}()
	_ = tmpLive.Close()

	liveDB, err := gorm.Open(sqlite.Open(tmpLive.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Initial sync (should create all tables)
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Initial sync failed: %v", err)
	}

	// Verify at least one table exists, e.g., cx_settings
	if !liveDB.Migrator().HasTable("cx_settings") {
		t.Error("Table cx_settings missing after initial sync")
	}

	// 2. Modify live DB (drop a column or change a table)
	// SQLite doesn't support DROP COLUMN easily, so let's DROP a table
	if err := liveDB.Exec("DROP TABLE auth_keys").Error; err != nil {
		t.Fatal(err)
	}

	// Sync again
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Second sync failed: %v", err)
	}

	if !liveDB.Migrator().HasTable("auth_keys") {
		t.Error("Table auth_keys missing after resync")
	}

	// 3. Test column addition
	// Drop a column by recreating the table without it
	if err := liveDB.Exec("ALTER TABLE cx_settings RENAME TO _cx_settings_old").Error; err != nil {
		t.Fatal(err)
	}
	// Create it without one column, e.g., smtp_host
	if err := liveDB.Exec("CREATE TABLE cx_settings (id INTEGER PRIMARY KEY, server_url TEXT)").Error; err != nil {
		t.Fatal(err)
	}
	// Insert some data
	if err := liveDB.Exec("INSERT INTO cx_settings (id, server_url) VALUES (1, 'http://localhost')").Error; err != nil {
		t.Fatal(err)
	}

	// Sync again
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Third sync failed: %v", err)
	}

	if !liveDB.Migrator().HasColumn("cx_settings", "smtp_host") {
		t.Error("Column smtp_host missing after resync")
	}

	// Verify data was preserved
	var serverURL string
	liveDB.Raw("SELECT server_url FROM cx_settings WHERE id = 1").Scan(&serverURL)
	if serverURL != "http://localhost" {
		t.Errorf("Data loss detected: expected http://localhost, got %s", serverURL)
	}

	// 4. Test index recreation
	// Drop and recreate an index with a different name or definition
	if err := liveDB.Exec("DROP INDEX idx_db_connections_driver").Error; err != nil {
		t.Fatal(err)
	}
	if err := liveDB.Exec("CREATE INDEX idx_db_connections_driver ON db_connections (driver, name)").Error; err != nil {
		t.Fatal(err)
	}

	// Sync again
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Fourth sync failed: %v", err)
	}

	// Check if index was restored to original definition (which is on 'driver' only)
	var indexSQL string
	liveDB.Raw("SELECT sql FROM sqlite_master WHERE type = 'index' AND name = 'idx_db_connections_driver'").Scan(&indexSQL)
	if !strings.Contains(strings.ToLower(indexSQL), "driver") || strings.Contains(strings.ToLower(indexSQL), "name") {
		t.Errorf("Index was not restored properly: %s", indexSQL)
	}
}

func TestSyncSchemaPruning(t *testing.T) {
	// Create a temporary live database
	tmpLive, err := os.CreateTemp("", "live-pruning-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(tmpLive.Name())
	}()
	_ = tmpLive.Close()

	liveDB, err := gorm.Open(sqlite.Open(tmpLive.Name()), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	// 1. Initial sync
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Initial sync failed: %v", err)
	}

	// 2. Add extra table 'foo'
	if err := liveDB.Exec("CREATE TABLE foo (id INTEGER PRIMARY KEY)").Error; err != nil {
		t.Fatal(err)
	}

	// 3. Add 'app_settings' table (should NOT be pruned)
	if err := liveDB.Exec("CREATE TABLE app_settings (key TEXT PRIMARY KEY, value TEXT)").Error; err != nil {
		t.Fatal(err)
	}

	// 4. Add extra column to 'cx_settings'
	if err := liveDB.Exec("ALTER TABLE cx_settings ADD COLUMN extra_col TEXT").Error; err != nil {
		t.Fatal(err)
	}

	// 5. Add extra index to 'cx_settings' (this table will be migrated, auto-pruning the index)
	if err := liveDB.Exec("CREATE INDEX idx_extra ON cx_settings (extra_col)").Error; err != nil {
		t.Fatal(err)
	}

	// 6. Add extra index to 'auth_keys' (this table will NOT be migrated, manual pruning needed)
	if err := liveDB.Exec("CREATE INDEX idx_extra_auth ON auth_keys (auth_key)").Error; err != nil {
		t.Fatal(err)
	}

	// 7. Sync again (should prune 'foo', 'extra_col', 'idx_extra', and 'idx_extra_auth', but KEEP 'app_settings')
	if err := SyncSchema(liveDB); err != nil {
		t.Fatalf("Sync with pruning failed: %v", err)
	}

	// Verify 'foo' is pruned
	if liveDB.Migrator().HasTable("foo") {
		t.Error("Table 'foo' was not pruned")
	}

	// Verify 'app_settings' is KEPT
	if !liveDB.Migrator().HasTable("app_settings") {
		t.Error("Table 'app_settings' was incorrectly pruned")
	}

	// Verify 'extra_col' is pruned from 'cx_settings'
	if liveDB.Migrator().HasColumn("cx_settings", "extra_col") {
		t.Error("Column 'extra_col' in 'cx_settings' was not pruned")
	}

	// Verify 'idx_extra' is pruned (auto-pruned by SQLite during table migration)
	var count int64
	liveDB.Raw("SELECT count(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_extra'").Scan(&count)
	if count > 0 {
		t.Error("Index 'idx_extra' was not pruned")
	}

	// Verify 'idx_extra_auth' is pruned (manually pruned by SyncSchema)
	liveDB.Raw("SELECT count(*) FROM sqlite_master WHERE type = 'index' AND name = 'idx_extra_auth'").Scan(&count)
	if count > 0 {
		t.Error("Index 'idx_extra_auth' was not pruned")
	}
}
