package main

import (
	"database/sql"
	"fmt"
	"time"
)

const (
	testerDBConnectionName      = "Tester SQLite"
	testerShellConnectionName   = "Tester Shell"
	testerAPIConnectionName     = "Tester API"
	testerWebhookConnectionName = "Tester Webhook"

	testerDBActionName      = "Tester DB Multi-Step"
	testerShellActionName   = "Tester Shell Multi-Step"
	testerWebActionName     = "Tester Web Multi-Step"
	testerWebhookActionName = "Tester Webhook Report"

	testerDBJobName      = "Tester DB Job"
	testerShellJobName   = "Tester Shell Job"
	testerWebJobName     = "Tester Web Job"
	testerWebhookJobName = "Tester Webhook Job"
)

type BootstrapOptions struct {
	Paths           AppPaths
	TesterHost      string
	APIPort         int
	WebhookPort     int
	TokenCommand    string
	ShellWorkingDir string
}

type dbStepFixture struct {
	order         int
	name          string
	sqlText       string
	expectation   any
	outputCapture any
}

type shellStepFixture struct {
	order                int
	name                 string
	runMode              string
	command              any
	scriptText           any
	shellPath            string
	workingDir           any
	outputCaptureMaxByte int
	outputTruncation     string
	expectation          any
	outputCapture        any
}

type webStepFixture struct {
	order           int
	name            string
	method          string
	url             string
	headers         any
	body            any
	expectation     any
	responseCapture any
}

func runBootstrap() error {
	paths, err := resolvePaths(CLI.DataDir)
	if err != nil {
		return err
	}

	return bootstrapChronixDB(CLI.Bootstrap.ChronixDB, BootstrapOptions{
		Paths:           paths,
		TesterHost:      CLI.Bootstrap.TesterHost,
		APIPort:         CLI.Bootstrap.APIPort,
		WebhookPort:     CLI.Bootstrap.WebhookPort,
		TokenCommand:    CLI.Bootstrap.TokenCommand,
		ShellWorkingDir: CLI.Bootstrap.ShellWorkingDir,
	})
}

func bootstrapChronixDB(chronixDBPath string, opts BootstrapOptions) error {
	store, err := OpenStore(opts.Paths)
	if err != nil {
		return err
	}
	if err := store.Close(); err != nil {
		return err
	}

	chDB, err := openSQLite(chronixDBPath)
	if err != nil {
		return fmt.Errorf("open chronix db: %w", err)
	}
	defer chDB.Close()

	tx, err := chDB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339Nano)
	tokenCommand := resolveTokenCommand(opts.Paths, opts.TokenCommand, opts.ShellWorkingDir)

	dbConnID, err := upsertDBConnection(tx, now, opts.Paths.TargetDB)
	if err != nil {
		return err
	}
	shellConnID, err := upsertShellConnection(tx, now)
	if err != nil {
		return err
	}
	webConnID, err := upsertWebTaskConnection(tx, now, testerAPIConnectionName, fmt.Sprintf("http://%s:%d", opts.TesterHost, opts.APIPort))
	if err != nil {
		return err
	}
	webhookConnID, err := upsertWebTaskConnection(tx, now, testerWebhookConnectionName, fmt.Sprintf("http://%s:%d", opts.TesterHost, opts.WebhookPort))
	if err != nil {
		return err
	}

	dbActionID, err := upsertAction(tx, now, testerDBActionName, "database", "Exercises multi-step SQLite execution, output capture, session persistence, and assertions.")
	if err != nil {
		return err
	}
	if err := replaceDBSteps(tx, dbActionID, []dbStepFixture{
		{
			order:       1,
			name:        "Create Session Temp Table",
			sqlText:     "CREATE TEMP TABLE IF NOT EXISTS tester_session (order_id INTEGER);",
			expectation: `{"kind":"noError"}`,
		},
		{
			order:       2,
			name:        "Create Order",
			sqlText:     "INSERT INTO orders (item_id, amount, status) VALUES (1, 10, 'pending');",
			expectation: `{"kind":"rowsAffected","op":"==","value":"1"}`,
		},
		{
			order:       3,
			name:        "Capture Session Order ID",
			sqlText:     "INSERT INTO tester_session (order_id) SELECT last_insert_rowid();",
			expectation: `{"kind":"rowsAffected","op":"==","value":"1"}`,
		},
		{
			order:         4,
			name:          "Read Captured Order ID",
			sqlText:       "SELECT order_id FROM tester_session ORDER BY rowid DESC LIMIT 1;",
			outputCapture: `{"order_id":{"source":"column","name":"order_id","row":"first"}}`,
		},
		{
			order:       5,
			name:        "Process Order",
			sqlText:     "UPDATE orders SET status = 'processed' WHERE id = {{order_id}};",
			expectation: `{"kind":"rowsAffected","op":"==","value":"1"}`,
		},
		{
			order:       6,
			name:        "Assert Order Status",
			sqlText:     "SELECT status FROM orders WHERE id = {{order_id}};",
			expectation: `{"kind":"fieldEqualsFirst","column":"status","expected":"processed"}`,
		},
	}); err != nil {
		return err
	}

	shellActionID, err := upsertAction(tx, now, testerShellActionName, "shell", "Exercises shell output capture, piping, line assertions, regex capture, and exit-code assertions.")
	if err != nil {
		return err
	}
	if err := replaceShellSteps(tx, shellActionID, []shellStepFixture{
		{
			order:                1,
			name:                 "Generate Token",
			runMode:              "command",
			command:              tokenCommand.Command,
			shellPath:            "/bin/sh",
			workingDir:           nullableString(tokenCommand.WorkingDir),
			outputCaptureMaxByte: 65536,
			outputTruncation:     "tail",
			outputCapture:        `{"token":{"source":"jsonpath","path":"$.token"}}`,
		},
		{
			order:            2,
			name:             "Echo Token",
			runMode:          "command",
			command:          `echo "Token is {{token}}"`,
			shellPath:        "/bin/sh",
			outputTruncation: "tail",
			expectation:      `{"kind":"contains","value":"Token is token-"}`,
		},
		{
			order:            3,
			name:             "First Line Expectation",
			runMode:          "command",
			command:          `printf 'first-line\nsecond-line\n'`,
			shellPath:        "/bin/sh",
			outputTruncation: "tail",
			expectation:      `{"kind":"firstLineEquals","value":"first-line"}`,
		},
		{
			order:                4,
			name:                 "Regex Capture",
			runMode:              "command",
			command:              `echo "User: dsherwin ID: 1234"`,
			shellPath:            "/bin/sh",
			outputCaptureMaxByte: 65536,
			outputTruncation:     "tail",
			outputCapture:        `{"user_id":{"source":"regex","pattern":"ID: (\\d+)","group":1}}`,
		},
		{
			order:            5,
			name:             "Echo User ID",
			runMode:          "command",
			command:          `echo "ID is {{user_id}}"`,
			shellPath:        "/bin/sh",
			outputTruncation: "tail",
			expectation:      `{"kind":"contains","value":"ID is 1234"}`,
		},
		{
			order:            6,
			name:             "Exit Code Assertion",
			runMode:          "command",
			command:          `exit 0`,
			shellPath:        "/bin/sh",
			outputTruncation: "tail",
			expectation:      `{"kind":"exitCodeEquals","value":"0"}`,
		},
	}); err != nil {
		return err
	}

	webActionID, err := upsertAction(tx, now, testerWebActionName, "webtask", "Exercises JSONPath, header capture, regex capture, and request piping against local tester HTTP fixtures.")
	if err != nil {
		return err
	}
	if err := replaceWebSteps(tx, webActionID, []webStepFixture{
		{
			order:           1,
			name:            "Get Fixture JSON",
			method:          "GET",
			url:             "/json",
			expectation:     `{"kind":"statusCode","op":"==","value":"200"}`,
			responseCapture: `{"fixture_title":{"source":"jsonpath","path":"$.slideshow.title"}}`,
		},
		{
			order:       2,
			name:        "Echo Fixture Title",
			method:      "POST",
			url:         "/echo",
			body:        `{"title":"{{fixture_title}}"}`,
			expectation: `{"kind":"bodyContains","value":"Chronix Fixture API"}`,
		},
		{
			order:           3,
			name:            "Get Response Header",
			method:          "GET",
			url:             "/response-headers?X-Test-ID=12345",
			expectation:     `{"kind":"statusCode","op":"==","value":"200"}`,
			responseCapture: `{"test_id":{"source":"header","name":"X-Test-ID"}}`,
		},
		{
			order:       4,
			name:        "Echo Header Value",
			method:      "POST",
			url:         "/echo",
			body:        `{"received_id":"{{test_id}}"}`,
			expectation: `{"kind":"bodyContains","value":"12345"}`,
		},
		{
			order:           5,
			name:            "Get Fixture HTML",
			method:          "GET",
			url:             "/html",
			expectation:     `{"kind":"statusCode","op":"==","value":"200"}`,
			responseCapture: `{"heading":{"source":"regex","pattern":"<h1>(.*?)</h1>","group":1}}`,
		},
		{
			order:       6,
			name:        "Echo HTML Heading",
			method:      "POST",
			url:         "/echo",
			body:        `{"heading":"{{heading}}"}`,
			expectation: `{"kind":"bodyContains","value":"Chronix Test Fixture"}`,
		},
	}); err != nil {
		return err
	}

	webhookActionID, err := upsertAction(tx, now, testerWebhookActionName, "webtask", "Sends a simple report payload to the tester webhook capture service.")
	if err != nil {
		return err
	}
	if err := replaceWebSteps(tx, webhookActionID, []webStepFixture{
		{
			order:       1,
			name:        "Send Report",
			method:      "POST",
			url:         "/report",
			body:        `{"report":"Chronix Test Success","source":"bootstrap"}`,
			expectation: `{"kind":"statusCode","op":"==","value":"200"}`,
		},
	}); err != nil {
		return err
	}

	if err := replaceTesterJobs(tx, now, dbConnID, shellConnID, webConnID, webhookConnID, dbActionID, shellActionID, webActionID, webhookActionID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	fmt.Printf("Bootstrapped Chronix DB at %s\n", chronixDBPath)
	fmt.Printf("Tester data dir: %s\n", opts.Paths.DataDir)
	fmt.Printf("Shell token command: %s\n", tokenCommand.Command)
	if tokenCommand.WorkingDir != nil && *tokenCommand.WorkingDir != "" {
		fmt.Printf("Shell token working dir: %s\n", *tokenCommand.WorkingDir)
	}

	return nil
}

func upsertDBConnection(tx *sql.Tx, now string, dsn string) (int64, error) {
	_, err := tx.Exec(`
INSERT INTO db_connections (name, driver, dsn, description, auto_check_enabled, auto_check_interval_seconds, created_at, updated_at, enabled, suspended)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	driver = excluded.driver,
	dsn = excluded.dsn,
	description = excluded.description,
	auto_check_enabled = excluded.auto_check_enabled,
	auto_check_interval_seconds = excluded.auto_check_interval_seconds,
	updated_at = excluded.updated_at,
	enabled = excluded.enabled,
	suspended = excluded.suspended
`, testerDBConnectionName, "sqlite", dsn, "SQLite target database managed by chronix-tester.", 0, 0, now, now, 1, 0)
	if err != nil {
		return 0, err
	}
	return lookupID(tx, "db_connections", testerDBConnectionName)
}

func upsertShellConnection(tx *sql.Tx, now string) (int64, error) {
	_, err := tx.Exec(`
INSERT INTO shell_connections (name, description, mode, auto_check_enabled, auto_check_interval_seconds, created_at, updated_at, enabled, suspended)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	description = excluded.description,
	mode = excluded.mode,
	auto_check_enabled = excluded.auto_check_enabled,
	auto_check_interval_seconds = excluded.auto_check_interval_seconds,
	updated_at = excluded.updated_at,
	enabled = excluded.enabled,
	suspended = excluded.suspended
`, testerShellConnectionName, "Local shell connection for chronix-tester shell fixtures.", "localhost", 0, 0, now, now, 1, 0)
	if err != nil {
		return 0, err
	}
	return lookupID(tx, "shell_connections", testerShellConnectionName)
}

func upsertWebTaskConnection(tx *sql.Tx, now string, name string, baseURL string) (int64, error) {
	_, err := tx.Exec(`
INSERT INTO webtask_connections (name, description, auth_type, base_url, auto_check_enabled, auto_check_interval_seconds, created_at, updated_at, enabled, suspended)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	description = excluded.description,
	auth_type = excluded.auth_type,
	base_url = excluded.base_url,
	auto_check_enabled = excluded.auto_check_enabled,
	auto_check_interval_seconds = excluded.auto_check_interval_seconds,
	updated_at = excluded.updated_at,
	enabled = excluded.enabled,
	suspended = excluded.suspended
`, name, "Local HTTP fixture service provided by chronix-tester.", "none", baseURL, 0, 0, now, now, 1, 0)
	if err != nil {
		return 0, err
	}
	return lookupID(tx, "webtask_connections", name)
}

func upsertAction(tx *sql.Tx, now string, name string, actionType string, description string) (int64, error) {
	_, err := tx.Exec(`
INSERT INTO actions (name, dialect, description, action_type, enabled, suspended, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(name) DO UPDATE SET
	dialect = excluded.dialect,
	description = excluded.description,
	action_type = excluded.action_type,
	updated_at = excluded.updated_at,
	enabled = excluded.enabled,
	suspended = excluded.suspended
`, name, "generic", description, actionType, 1, 0, now, now)
	if err != nil {
		return 0, err
	}
	return lookupID(tx, "actions", name)
}

func replaceDBSteps(tx *sql.Tx, actionID int64, steps []dbStepFixture) error {
	if _, err := tx.Exec(`DELETE FROM action_steps WHERE action_id = ?`, actionID); err != nil {
		return err
	}

	for _, step := range steps {
		if _, err := tx.Exec(`
INSERT INTO action_steps (action_id, step_order, name, sql_text, expectation, output_capture)
VALUES (?, ?, ?, ?, ?, ?)`,
			actionID, step.order, step.name, step.sqlText, step.expectation, step.outputCapture); err != nil {
			return err
		}
	}

	return nil
}

func replaceShellSteps(tx *sql.Tx, actionID int64, steps []shellStepFixture) error {
	if _, err := tx.Exec(`DELETE FROM shell_action_steps WHERE action_id = ?`, actionID); err != nil {
		return err
	}

	for _, step := range steps {
		maxBytes := step.outputCaptureMaxByte
		if maxBytes == 0 {
			maxBytes = 65536
		}
		truncation := step.outputTruncation
		if truncation == "" {
			truncation = "tail"
		}

		if _, err := tx.Exec(`
INSERT INTO shell_action_steps (
	action_id, step_order, name, run_mode, command, script_text, shell_path, working_dir,
	output_capture_max_bytes, output_truncation, expectation, output_capture
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			actionID,
			step.order,
			step.name,
			step.runMode,
			step.command,
			step.scriptText,
			step.shellPath,
			step.workingDir,
			maxBytes,
			truncation,
			step.expectation,
			step.outputCapture,
		); err != nil {
			return err
		}
	}

	return nil
}

func replaceWebSteps(tx *sql.Tx, actionID int64, steps []webStepFixture) error {
	if _, err := tx.Exec(`DELETE FROM webtask_action_steps WHERE action_id = ?`, actionID); err != nil {
		return err
	}

	for _, step := range steps {
		if _, err := tx.Exec(`
INSERT INTO webtask_action_steps (action_id, step_order, name, method, url, headers, body, expectation, response_capture)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			actionID, step.order, step.name, step.method, step.url, step.headers, step.body, step.expectation, step.responseCapture); err != nil {
			return err
		}
	}

	return nil
}

func replaceTesterJobs(tx *sql.Tx, now string, dbConnID int64, shellConnID int64, webConnID int64, webhookConnID int64, dbActionID int64, shellActionID int64, webActionID int64, webhookActionID int64) error {
	if _, err := tx.Exec(`DELETE FROM jobs WHERE name IN (?, ?, ?, ?)`, testerDBJobName, testerShellJobName, testerWebJobName, testerWebhookJobName); err != nil {
		return err
	}

	manualSchedule := `{"kind":"manual"}`

	inserts := []struct {
		name                string
		description         string
		connectionID        any
		shellConnectionID   any
		webtaskConnectionID any
		actionID            int64
		targetKind          string
	}{
		{
			name:         testerDBJobName,
			description:  "Runs the SQLite multi-step tester action.",
			connectionID: dbConnID,
			actionID:     dbActionID,
			targetKind:   "database",
		},
		{
			name:              testerShellJobName,
			description:       "Runs the shell multi-step tester action.",
			shellConnectionID: shellConnID,
			actionID:          shellActionID,
			targetKind:        "shell",
		},
		{
			name:                testerWebJobName,
			description:         "Runs the web-task multi-step tester action.",
			webtaskConnectionID: webConnID,
			actionID:            webActionID,
			targetKind:          "webtask",
		},
		{
			name:                testerWebhookJobName,
			description:         "Sends a report to the tester webhook capture endpoint.",
			webtaskConnectionID: webhookConnID,
			actionID:            webhookActionID,
			targetKind:          "webtask",
		},
	}

	for _, job := range inserts {
		if _, err := tx.Exec(`
INSERT INTO jobs (
	name, description, connection_id, shell_connection_id, webtask_connection_id,
	action_id, target_kind, schedule_json, enabled, suspended, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			job.name,
			job.description,
			job.connectionID,
			job.shellConnectionID,
			job.webtaskConnectionID,
			job.actionID,
			job.targetKind,
			manualSchedule,
			1,
			0,
			now,
			now,
		); err != nil {
			return err
		}
	}

	return nil
}

func lookupID(tx *sql.Tx, table string, name string) (int64, error) {
	var id int64
	err := tx.QueryRow(fmt.Sprintf(`SELECT id FROM %s WHERE name = ?`, table), name).Scan(&id)
	return id, err
}

func nullableString(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
