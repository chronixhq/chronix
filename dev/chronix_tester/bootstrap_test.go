package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/glebarez/go-sqlite"
)

func TestBootstrapChronixDBIsIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	chronixDBPath := filepath.Join(tempDir, "chronix.db")
	createChronixFixtureDB(t, chronixDBPath)

	paths := AppPaths{
		DataDir:   filepath.Join(tempDir, "tester"),
		ResultsDB: filepath.Join(tempDir, "tester", "results.db"),
		TargetDB:  filepath.Join(tempDir, "tester", "target.db"),
	}

	opts := BootstrapOptions{
		Paths:           paths,
		TesterHost:      "localhost",
		APIPort:         5181,
		WebhookPort:     5182,
		TokenCommand:    "/tmp/chronix-tester generate-token --data-dir /tmp/chronix-tester",
		ShellWorkingDir: "/tmp",
	}

	if err := bootstrapChronixDB(chronixDBPath, opts); err != nil {
		t.Fatalf("first bootstrap failed: %v", err)
	}
	if err := bootstrapChronixDB(chronixDBPath, opts); err != nil {
		t.Fatalf("second bootstrap failed: %v", err)
	}

	db, err := sql.Open("sqlite", chronixDBPath)
	if err != nil {
		t.Fatalf("open chronix db: %v", err)
	}
	defer db.Close()

	assertCount(t, db, `SELECT COUNT(*) FROM jobs WHERE name LIKE 'Tester %'`, 4)
	assertCount(t, db, `SELECT COUNT(*) FROM action_steps WHERE action_id = (SELECT id FROM actions WHERE name = ?)`, 6, testerDBActionName)
	assertCount(t, db, `SELECT COUNT(*) FROM shell_action_steps WHERE action_id = (SELECT id FROM actions WHERE name = ?)`, 6, testerShellActionName)
	assertCount(t, db, `SELECT COUNT(*) FROM webtask_action_steps WHERE action_id = (SELECT id FROM actions WHERE name = ?)`, 6, testerWebActionName)

	var dsn string
	if err := db.QueryRow(`SELECT dsn FROM db_connections WHERE name = ?`, testerDBConnectionName).Scan(&dsn); err != nil {
		t.Fatalf("lookup db connection: %v", err)
	}
	if dsn != paths.TargetDB {
		t.Fatalf("unexpected target db dsn: got %q want %q", dsn, paths.TargetDB)
	}

	var command string
	var workingDir sql.NullString
	if err := db.QueryRow(`
SELECT command, working_dir
FROM shell_action_steps
WHERE action_id = (SELECT id FROM actions WHERE name = ?)
  AND step_order = 1`, testerShellActionName).Scan(&command, &workingDir); err != nil {
		t.Fatalf("lookup shell step: %v", err)
	}
	if command != opts.TokenCommand {
		t.Fatalf("unexpected token command: got %q want %q", command, opts.TokenCommand)
	}
	if !workingDir.Valid || workingDir.String != opts.ShellWorkingDir {
		t.Fatalf("unexpected working dir: got %q want %q", workingDir.String, opts.ShellWorkingDir)
	}
}

func createChronixFixtureDB(t *testing.T, path string) {
	t.Helper()

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open fixture db: %v", err)
	}
	defer db.Close()

	schema := `
CREATE TABLE db_connections (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	driver TEXT NOT NULL,
	dsn TEXT NOT NULL,
	description TEXT,
	auto_check_enabled INTEGER NOT NULL,
	auto_check_interval_seconds INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	suspended INTEGER NOT NULL
);
CREATE TABLE shell_connections (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	mode TEXT NOT NULL,
	auto_check_enabled INTEGER NOT NULL,
	auto_check_interval_seconds INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	suspended INTEGER NOT NULL
);
CREATE TABLE webtask_connections (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	description TEXT,
	auth_type TEXT NOT NULL,
	base_url TEXT,
	auto_check_enabled INTEGER NOT NULL,
	auto_check_interval_seconds INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	suspended INTEGER NOT NULL
);
CREATE TABLE actions (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	dialect TEXT NOT NULL,
	description TEXT,
	action_type TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	suspended INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);
CREATE TABLE action_steps (
	id INTEGER PRIMARY KEY,
	action_id INTEGER NOT NULL,
	step_order INTEGER NOT NULL,
	name TEXT NOT NULL,
	sql_text TEXT NOT NULL,
	expectation TEXT,
	output_capture TEXT
);
CREATE TABLE shell_action_steps (
	id INTEGER PRIMARY KEY,
	action_id INTEGER NOT NULL,
	step_order INTEGER NOT NULL,
	name TEXT NOT NULL,
	run_mode TEXT NOT NULL,
	command TEXT,
	script_text TEXT,
	shell_path TEXT NOT NULL,
	working_dir TEXT,
	output_capture_max_bytes INTEGER NOT NULL,
	output_truncation TEXT NOT NULL,
	expectation TEXT,
	output_capture TEXT
);
CREATE TABLE webtask_action_steps (
	id INTEGER PRIMARY KEY,
	action_id INTEGER NOT NULL,
	step_order INTEGER NOT NULL,
	name TEXT NOT NULL,
	method TEXT NOT NULL,
	url TEXT NOT NULL,
	headers TEXT,
	body TEXT,
	expectation TEXT,
	response_capture TEXT
);
CREATE TABLE jobs (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL,
	description TEXT,
	connection_id INTEGER,
	shell_connection_id INTEGER,
	webtask_connection_id INTEGER,
	action_id INTEGER NOT NULL,
	target_kind TEXT NOT NULL,
	schedule_json TEXT NOT NULL,
	enabled INTEGER NOT NULL,
	suspended INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create fixture schema: %v", err)
	}
}

func assertCount(t *testing.T, db *sql.DB, query string, want int, args ...any) {
	t.Helper()

	var got int
	if err := db.QueryRow(query, args...).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}
