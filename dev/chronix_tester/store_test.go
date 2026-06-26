package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenStoreCapturesDetailedTargetActivity(t *testing.T) {
	dataDir := t.TempDir()
	paths := AppPaths{
		DataDir:   dataDir,
		ResultsDB: filepath.Join(dataDir, "results.db"),
		TargetDB:  filepath.Join(dataDir, "target.db"),
	}

	store, err := OpenStore(paths)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if _, err := store.Target.Exec(`INSERT INTO inventory (name, quantity) VALUES ('Fixture Widget', 7)`); err != nil {
		t.Fatalf("insert inventory: %v", err)
	}
	if _, err := store.Target.Exec(`UPDATE inventory SET quantity = 8 WHERE name = 'Fixture Widget'`); err != nil {
		t.Fatalf("update inventory: %v", err)
	}
	if _, err := store.Target.Exec(`DELETE FROM inventory WHERE name = 'Fixture Widget'`); err != nil {
		t.Fatalf("delete inventory: %v", err)
	}

	snapshot, err := store.LoadSnapshot(10)
	if err != nil {
		t.Fatalf("load snapshot: %v", err)
	}

	if len(snapshot.TargetActivity) < 3 {
		t.Fatalf("expected at least 3 target activity entries, got %d", len(snapshot.TargetActivity))
	}

	joined := snapshot.TargetActivity[0].OldData + snapshot.TargetActivity[1].NewData + snapshot.TargetActivity[2].NewData
	if !strings.Contains(joined, "Fixture Widget") {
		t.Fatalf("expected trigger payloads to include row details, got %q", joined)
	}
}
