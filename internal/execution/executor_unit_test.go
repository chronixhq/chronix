package execution

import (
	"encoding/json"
	"testing"
)

func TestIntFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected int
	}{
		{"float64", float64(1), 1},
		{"int", int(2), 2},
		{"int64", int64(3), 3},
		{"string", "4", 4},
		{"string invalid", "abc", 0},
		{"nil", nil, 0},
		{"json.Number", json.Number("5"), 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intFromAny(tt.input)
			if got != tt.expected {
				t.Errorf("intFromAny() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEvaluateExpectation(t *testing.T) {
	vars := map[string]any{"count": "5"}

	t.Run("none", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "none"}, 0, 0, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("noError", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "noError"}, 0, 0, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("rowExists success", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "rowExists"}, 1, 0, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("rowExists failure", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "rowExists"}, 0, 0, nil, vars)
		if pass {
			t.Error("expected failure")
		}
	})

	t.Run("noRowsReturned success", func(t *testing.T) {
		pass, _, meta, _ := evaluateExpectation(map[string]any{"kind": "noRowsReturned"}, 0, 0, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
		if meta["expect_kind"] != "noRowsReturned" {
			t.Errorf("expected expect_kind noRowsReturned, got %#v", meta["expect_kind"])
		}
	})

	t.Run("noRowsReturned failure", func(t *testing.T) {
		pass, msg, _, _ := evaluateExpectation(map[string]any{"kind": "noRowsReturned"}, 2, 0, []map[string]any{{"x": 1}, {"x": 2}}, vars)
		if pass {
			t.Error("expected failure")
		}
		if msg == "" || !contains(msg, "Zero rows were expected") {
			t.Errorf("expected helpful message, got %q", msg)
		}
	})

	t.Run("fieldEqualsFirst success", func(t *testing.T) {
		results := []map[string]any{{"status": "OK"}, {"status": "FAIL"}}
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "fieldEqualsFirst", "column": "status", "expected": "OK"}, 2, 0, results, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("fieldEqualsLast success", func(t *testing.T) {
		results := []map[string]any{{"status": "OK"}, {"status": "FAIL"}}
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "fieldEqualsLast", "column": "status", "expected": "FAIL"}, 2, 0, results, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("rowsAffected success", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "rowsAffected", "op": ">=", "value": "1"}, 0, 1, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("rowsAffected variable success", func(t *testing.T) {
		pass, _, _, _ := evaluateExpectation(map[string]any{"kind": "rowsAffected", "op": "==", "value": "${count}"}, 0, 5, nil, vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("rowsAffected negative failure", func(t *testing.T) {
		pass, msg, _, _ := evaluateExpectation(map[string]any{"kind": "rowsAffected", "op": "==", "value": "1"}, 0, -1, nil, vars)
		if pass {
			t.Error("expected failure")
		}
		if msg == "" || !contains(msg, "did not report") {
			t.Errorf("expected helpful message, got %q", msg)
		}
	})
}

func contains(s, substr string) bool {
	return (s != "" && substr != "" && (len(s) >= len(substr))) && (s == substr || (len(substr) == 0) || (s != "" && (len(s) >= len(substr)) && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}

func TestEvaluateShellExpectation(t *testing.T) {
	vars := map[string]any{"exit": "0", "target": "world"}

	t.Run("noError success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "noError"}, 0, nil, nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("exitCodeEquals success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "exitCodeEquals", "value": "${exit}"}, 0, nil, nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("contains success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "contains", "value": "hello ${target}"}, 0, []byte("hello world"), nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("regexMatch success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "regexMatch", "value": "h[e]llo"}, 0, []byte("hello"), nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("regexMatch group success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "regexMatch", "value": "ID: (\\d+)", "group": "1", "expected": "1234"}, 0, []byte("User: dsherwin ID: 1234"), nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})

	t.Run("regexMatch group failure", func(t *testing.T) {
		pass, msg, _ := evaluateShellExpectation(map[string]any{"kind": "regexMatch", "value": "ID: (\\d+)", "group": "1", "expected": "5678"}, 0, []byte("User: dsherwin ID: 1234"), nil, false, false, "tail", vars)
		if pass {
			t.Error("expected failure")
		}
		if !contains(msg, "expected \"5678\", got \"1234\"") {
			t.Errorf("expected error message about mismatch, got %q", msg)
		}
	})

	t.Run("firstLineEquals success", func(t *testing.T) {
		pass, _, _ := evaluateShellExpectation(map[string]any{"kind": "firstLineEquals", "value": "line1"}, 0, []byte("line1\nline2"), nil, false, false, "tail", vars)
		if !pass {
			t.Error("expected pass")
		}
	})
}

func TestBindParams(t *testing.T) {
	vars := map[string]any{"id": 123, "name": "alice"}

	t.Run("postgres placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}} AND name = {{name}}"
		gotSQL, gotArgs, err := bindParams("postgres", sql, vars)
		if err != nil {
			t.Fatal(err)
		}
		expectedSQL := "SELECT * FROM users WHERE id = $1 AND name = $2"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
		if len(gotArgs) != 2 || gotArgs[0] != 123 || gotArgs[1] != "alice" {
			t.Errorf("unexpected args: %v", gotArgs)
		}
	})

	t.Run("mysql placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}} AND name = {{name}}"
		gotSQL, _, err := bindParams("mysql", sql, vars)
		if err != nil {
			t.Fatal(err)
		}
		expectedSQL := "SELECT * FROM users WHERE id = ? AND name = ?"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
	})

	t.Run("sqlite placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}}"
		gotSQL, _, _ := bindParams("sqlite", sql, vars)
		expectedSQL := "SELECT * FROM users WHERE id = ?"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
	})

	t.Run("sqlserver placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}} AND name = {{name}}"
		gotSQL, _, _ := bindParams("mssql", sql, vars)
		expectedSQL := "SELECT * FROM users WHERE id = @p1 AND name = @p2"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
	})

	t.Run("oracle placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}} AND name = {{name}}"
		gotSQL, _, _ := bindParams("oracle", sql, vars)
		expectedSQL := "SELECT * FROM users WHERE id = :1 AND name = :2"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
	})

	t.Run("snowflake placeholders", func(t *testing.T) {
		sql := "SELECT * FROM users WHERE id = {{id}}"
		gotSQL, _, _ := bindParams("snowflake", sql, vars)
		expectedSQL := "SELECT * FROM users WHERE id = ?"
		if gotSQL != expectedSQL {
			t.Errorf("expected %q, got %q", expectedSQL, gotSQL)
		}
	})
}
