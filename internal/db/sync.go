package db

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"gorm.io/gorm"
)

type schemaItem struct {
	Type string
	Name string
	SQL  string
}

// SyncSchema verifies the schema of the live database against the embedded schema.db.
// It creates missing tables/indexes and updates existing ones if their definitions differ.
func SyncSchema(liveDB *gorm.DB) error {
	slog.Debug("Starting database schema verification")

	// 1. Write EmbeddedSchemaDB to a temp file
	tmpFile, err := os.CreateTemp("", "chronix-schema-*.db")
	if err != nil {
		return fmt.Errorf("create temp schema file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(EmbeddedSchemaDB); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("write temp schema file: %w", err)
	}
	_ = tmpFile.Close()

	// 2. Attach reference DB
	// We use a raw connection to avoid GORM's overhead for these meta-operations
	if err := liveDB.Exec("ATTACH DATABASE ? AS ref", tmpFile.Name()).Error; err != nil {
		return fmt.Errorf("attach reference database: %w", err)
	}
	defer func() {
		_ = liveDB.Exec("DETACH DATABASE ref")
	}()

	// 3. Get all items from reference DB
	var refItems []schemaItem
	if err := liveDB.Raw("SELECT type, name, sql FROM ref.sqlite_master WHERE name NOT LIKE 'sqlite_%' AND sql IS NOT NULL").Scan(&refItems).Error; err != nil {
		return fmt.Errorf("query reference schema: %w", err)
	}

	// 4. Disable foreign keys temporarily for the migration
	var originalFKSetting int
	liveDB.Raw("PRAGMA foreign_keys").Scan(&originalFKSetting)
	liveDB.Exec("PRAGMA foreign_keys = OFF")
	defer func() {
		if originalFKSetting == 1 {
			liveDB.Exec("PRAGMA foreign_keys = ON")
		}
	}()

	for _, item := range refItems {
		var liveSQL string
		err := liveDB.Raw("SELECT sql FROM main.sqlite_master WHERE type = ? AND name = ?", item.Type, item.Name).Scan(&liveSQL).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("check live schema for %s %s: %w", item.Type, item.Name, err)
		}

		if liveSQL == "" {
			// Item missing in live DB
			slog.Info("Creating missing schema item", slog.String("type", item.Type), slog.String("name", item.Name))
			if err := liveDB.Exec(item.SQL).Error; err != nil {
				return fmt.Errorf("create %s %s: %w", item.Type, item.Name, err)
			}
			continue
		}

		// Normalize SQL for comparison (basic cleanup)
		if normalizeSQL(liveSQL) == normalizeSQL(item.SQL) {
			continue
		}

		// Definitions differ
		slog.Warn("Schema mismatch detected", slog.String("type", item.Type), slog.String("name", item.Name))

		switch item.Type {
		case "table":
			if err := syncTable(liveDB, item.Name, item.SQL); err != nil {
				return fmt.Errorf("sync table %s: %w", item.Name, err)
			}
		case "index":
			slog.Info("Recreating index", slog.String("name", item.Name))
			if err := liveDB.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", item.Name)).Error; err != nil {
				return fmt.Errorf("drop index %s: %w", item.Name, err)
			}
			if err := liveDB.Exec(item.SQL).Error; err != nil {
				return fmt.Errorf("recreate index %s: %w", item.Name, err)
			}
		}
	}

	// 5. Prune extra items (tables/indexes) in live DB that are not in reference DB
	type liveSchemaItem struct {
		Type    string
		Name    string
		TblName string `gorm:"column:tbl_name"`
	}
	var liveItems []liveSchemaItem
	if err := liveDB.Raw("SELECT type, name, tbl_name FROM main.sqlite_master WHERE name NOT LIKE 'sqlite_%' AND name != 'app_settings' AND tbl_name != 'app_settings' AND sql IS NOT NULL").Scan(&liveItems).Error; err != nil {
		return fmt.Errorf("query live schema for pruning: %w", err)
	}

	refItemNames := make(map[string]bool)
	for _, item := range refItems {
		refItemNames[item.Name] = true
	}

	for _, liveItem := range liveItems {
		if !refItemNames[liveItem.Name] {
			slog.Warn("Pruning extra schema item", slog.String("type", liveItem.Type), slog.String("name", liveItem.Name))
			dropSQL := ""
			switch liveItem.Type {
			case "table":
				dropSQL = fmt.Sprintf("DROP TABLE %s", liveItem.Name)
			case "index":
				dropSQL = fmt.Sprintf("DROP INDEX %s", liveItem.Name)
			case "view":
				dropSQL = fmt.Sprintf("DROP VIEW %s", liveItem.Name)
			case "trigger":
				dropSQL = fmt.Sprintf("DROP TRIGGER %s", liveItem.Name)
			}
			if dropSQL != "" {
				if err := liveDB.Exec(dropSQL).Error; err != nil {
					return fmt.Errorf("prune %s %s: %w", liveItem.Type, liveItem.Name, err)
				}
			}
		}
	}

	slog.Debug("Database schema verification complete")
	return nil
}

func normalizeSQL(sql string) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	for strings.Contains(sql, "  ") {
		sql = strings.ReplaceAll(sql, "  ", " ")
	}
	return strings.TrimSpace(strings.ToLower(sql))
}

func syncTable(db *gorm.DB, name string, refSQL string) error {
	slog.Info("Migrating table", slog.String("table", name))

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. Rename existing table
		tempName := fmt.Sprintf("_%s_old", name)
		_ = tx.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tempName)) // Cleanup if previous attempt failed
		if err := tx.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", name, tempName)).Error; err != nil {
			return err
		}

		// 2. Create new table with correct schema
		if err := tx.Exec(refSQL).Error; err != nil {
			return err
		}

		// 3. Copy data for columns
		rows, err := tx.Raw(fmt.Sprintf("PRAGMA table_info(%s)", tempName)).Rows()
		if err != nil {
			return err
		}
		defer func() {
			_ = rows.Close()
		}()

		oldCols := make(map[string]bool)
		for rows.Next() {
			var cid int
			var cname, ctype string
			var notnull, pk int
			var dflt any
			if err := rows.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
				return err
			}
			oldCols[strings.ToLower(cname)] = true
		}

		rows2, err := tx.Raw(fmt.Sprintf("PRAGMA table_info(%s)", name)).Rows()
		if err != nil {
			return err
		}
		defer func() {
			_ = rows2.Close()
		}()

		var targetCols []string
		var selectExprs []string
		for rows2.Next() {
			var cid int
			var cname, ctype string
			var notnull, pk int
			var dflt any
			if err := rows2.Scan(&cid, &cname, &ctype, &notnull, &dflt, &pk); err != nil {
				return err
			}
			if oldCols[strings.ToLower(cname)] {
				targetCols = append(targetCols, cname)
				selectExprs = append(selectExprs, cname)
			} else if notnull == 1 && dflt == nil {
				// Column is new, NOT NULL, and has no DEFAULT in schema.
				// We must provide a value to avoid "NOT NULL constraint failed".
				targetCols = append(targetCols, cname)

				lowName := strings.ToLower(cname)
				lowType := strings.ToLower(ctype)
				if strings.Contains(lowType, "bool") || strings.Contains(lowType, "int") {
					val := "1"
					if strings.Contains(lowName, "suspended") {
						val = "0"
					}
					selectExprs = append(selectExprs, val) // Assume enabled/active by default for new columns, except suspended
				} else if strings.Contains(lowType, "timestamp") || strings.Contains(lowType, "datetime") {
					selectExprs = append(selectExprs, "CURRENT_TIMESTAMP")
				} else {
					selectExprs = append(selectExprs, "''")
				}
			}
		}

		if len(targetCols) > 0 {
			cols := strings.Join(targetCols, ", ")
			exprs := strings.Join(selectExprs, ", ")
			copySQL := fmt.Sprintf("INSERT INTO %s (%s) SELECT %s FROM %s", name, cols, exprs, tempName)
			if err := tx.Exec(copySQL).Error; err != nil {
				return err
			}
		}

		// 4. Drop old table
		if err := tx.Exec(fmt.Sprintf("DROP TABLE %s", tempName)).Error; err != nil {
			return err
		}

		return nil
	})
}
