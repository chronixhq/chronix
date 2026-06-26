package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	_ "github.com/glebarez/go-sqlite"
)

type Store struct {
	Paths   AppPaths
	Results *sql.DB
	Target  *sql.DB
}

const resultsSchema = `
CREATE TABLE IF NOT EXISTS imap_settings (
	id INTEGER PRIMARY KEY,
	host TEXT,
	port INTEGER,
	user TEXT,
	pass TEXT,
	ssl BOOLEAN
);

CREATE TABLE IF NOT EXISTS api_logs (
	id INTEGER PRIMARY KEY,
	method TEXT,
	path TEXT,
	headers TEXT,
	body TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS webhook_logs (
	id INTEGER PRIMARY KEY,
	method TEXT,
	path TEXT,
	headers TEXT,
	body TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS imap_logs (
	id INTEGER PRIMARY KEY,
	subject TEXT,
	from_addr TEXT,
	body TEXT,
	received_at TIMESTAMP,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shell_logs (
	id INTEGER PRIMARY KEY,
	args TEXT,
	output TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_api_logs_created_at ON api_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhook_logs_created_at ON webhook_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_imap_logs_created_at ON imap_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_shell_logs_created_at ON shell_logs (created_at DESC);
`

const targetSchema = `
CREATE TABLE IF NOT EXISTS inventory (
	id INTEGER PRIMARY KEY,
	name TEXT,
	quantity INTEGER
);

CREATE TABLE IF NOT EXISTS orders (
	id INTEGER PRIMARY KEY,
	item_id INTEGER,
	amount INTEGER,
	status TEXT
);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	username TEXT,
	email TEXT
);

CREATE TABLE IF NOT EXISTS target_activity (
	id INTEGER PRIMARY KEY,
	table_name TEXT,
	operation TEXT,
	old_data TEXT,
	new_data TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_target_activity_created_at ON target_activity (created_at DESC);
`

func OpenStore(paths AppPaths) (*Store, error) {
	if err := os.MkdirAll(paths.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir %s: %w", paths.DataDir, err)
	}

	resultsDB, err := openSQLite(paths.ResultsDB)
	if err != nil {
		return nil, fmt.Errorf("open results db: %w", err)
	}

	targetDB, err := openSQLite(paths.TargetDB)
	if err != nil {
		_ = resultsDB.Close()
		return nil, fmt.Errorf("open target db: %w", err)
	}

	store := &Store{
		Paths:   paths,
		Results: resultsDB,
		Target:  targetDB,
	}

	if err := initResultsSchema(resultsDB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("init results schema: %w", err)
	}
	if err := initTargetSchema(targetDB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("init target schema: %w", err)
	}
	if err := seedTargetData(targetDB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("seed target data: %w", err)
	}
	if err := setupTriggers(targetDB); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("setup triggers: %w", err)
	}

	return store, nil
}

func (s *Store) Close() error {
	var errs []error
	if s.Results != nil {
		errs = append(errs, s.Results.Close())
	}
	if s.Target != nil {
		errs = append(errs, s.Target.Close())
	}
	return errors.Join(errs...)
}

func ResetStore(paths AppPaths) error {
	for _, dbPath := range []string{paths.ResultsDB, paths.TargetDB} {
		if err := os.Remove(dbPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("remove %s: %w", dbPath, err)
		}
	}

	store, err := OpenStore(paths)
	if err != nil {
		return err
	}
	return store.Close()
}

func openSQLite(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func initResultsSchema(db *sql.DB) error {
	if _, err := db.Exec(resultsSchema); err != nil {
		return err
	}
	if err := ensureColumn(db, "webhook_logs", "method", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumn(db, "webhook_logs", "path", "TEXT"); err != nil {
		return err
	}
	return nil
}

func initTargetSchema(db *sql.DB) error {
	_, err := db.Exec(targetSchema)
	return err
}

func ensureColumn(db *sql.DB, table string, column string, columnDDL string) error {
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}

	_, err = db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN %s %s`, table, column, columnDDL))
	return err
}

func seedTargetData(db *sql.DB) error {
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM inventory`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.Exec(`INSERT INTO inventory (name, quantity) VALUES ('Widget A', 100), ('Gadget B', 50)`); err != nil {
			return err
		}
	}

	if err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		if _, err := db.Exec(`INSERT INTO users (username, email) VALUES ('chronix-admin', 'admin@example.com')`); err != nil {
			return err
		}
	}

	return nil
}

func setupTriggers(db *sql.DB) error {
	type payload struct {
		oldJSON string
		newJSON string
	}

	tables := map[string]payload{
		"inventory": {
			oldJSON: `json_object('id', OLD.id, 'name', OLD.name, 'quantity', OLD.quantity)`,
			newJSON: `json_object('id', NEW.id, 'name', NEW.name, 'quantity', NEW.quantity)`,
		},
		"orders": {
			oldJSON: `json_object('id', OLD.id, 'item_id', OLD.item_id, 'amount', OLD.amount, 'status', OLD.status)`,
			newJSON: `json_object('id', NEW.id, 'item_id', NEW.item_id, 'amount', NEW.amount, 'status', NEW.status)`,
		},
		"users": {
			oldJSON: `json_object('id', OLD.id, 'username', OLD.username, 'email', OLD.email)`,
			newJSON: `json_object('id', NEW.id, 'username', NEW.username, 'email', NEW.email)`,
		},
	}

	for table, p := range tables {
		stmts := []string{
			fmt.Sprintf(`DROP TRIGGER IF EXISTS trg_%s_insert`, table),
			fmt.Sprintf(`DROP TRIGGER IF EXISTS trg_%s_update`, table),
			fmt.Sprintf(`DROP TRIGGER IF EXISTS trg_%s_delete`, table),
			fmt.Sprintf(`
CREATE TRIGGER trg_%s_insert AFTER INSERT ON %s
BEGIN
	INSERT INTO target_activity (table_name, operation, new_data)
	VALUES ('%s', 'INSERT', %s);
END;`, table, table, table, p.newJSON),
			fmt.Sprintf(`
CREATE TRIGGER trg_%s_update AFTER UPDATE ON %s
BEGIN
	INSERT INTO target_activity (table_name, operation, old_data, new_data)
	VALUES ('%s', 'UPDATE', %s, %s);
END;`, table, table, table, p.oldJSON, p.newJSON),
			fmt.Sprintf(`
CREATE TRIGGER trg_%s_delete AFTER DELETE ON %s
BEGIN
	INSERT INTO target_activity (table_name, operation, old_data)
	VALUES ('%s', 'DELETE', %s);
END;`, table, table, table, p.oldJSON),
		}

		for _, stmt := range stmts {
			if _, err := db.Exec(stmt); err != nil {
				return fmt.Errorf("%s trigger: %w", table, err)
			}
		}
	}

	return nil
}
