package db

import (
	_ "embed"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
)

// embeddedSchemaDB contains the SQLite database file with the application schema.
// The path is relative to this source file directory.
//
//go:embed assets/schema.db
var EmbeddedSchemaDB []byte

// EnsureSQLiteFile checks whether a SQLite database exists at dbPath. If it does not,
// it creates the file from the embedded schema database. It returns true if the file
// was created during this call.
func EnsureSQLiteFile(dbPath string) (bool, error) {
	if dbPath == "" {
		return false, errors.New("dbPath is empty")
	}

	if _, err := os.Stat(dbPath); err == nil {
		// Already exists
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	// Ensure parent directory exists with reasonably restrictive permissions.
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o770); err != nil { // directory perms can be adjusted by caller if needed
		return false, err
	}

	// Write to a temp file then atomically rename into place.
	tmp := dbPath + ".tmp"
	if err := os.WriteFile(tmp, EmbeddedSchemaDB, 0o660); err != nil {
		return false, err
	}
	if err := os.Rename(tmp, dbPath); err != nil {
		_ = os.Remove(tmp)
		return false, err
	}

	slog.Info("Created SQLite database from embedded schema", slog.String("path", dbPath))
	return true, nil
}
