package main

import (
	"reflect"
	"testing"
)

func TestSafeStr(t *testing.T) {
	s := "hello"
	if got := safeStr(&s); got != "hello" {
		t.Errorf("safeStr(&s) = %q; want %q", got, "hello")
	}
	if got := safeStr(nil); got != "" {
		t.Errorf("safeStr(nil) = %q; want %q", got, "")
	}
}

func TestSafeInt(t *testing.T) {
	i := 42
	if got := safeInt(&i, 0); got != 42 {
		t.Errorf("safeInt(&i, 0) = %d; want %d", got, 42)
	}
	if got := safeInt(nil, 10); got != 10 {
		t.Errorf("safeInt(nil, 10) = %d; want %d", got, 10)
	}
}

func TestParseServerFlag(t *testing.T) {
	tests := []struct {
		in       string
		wantHost string
		wantPort int
	}{
		{"", "localhost", 5172},
		{"1.2.3.4", "1.2.3.4", 5172},
		{"myserver:6060", "myserver", 6060},
		{"localhost:abc", "localhost", 5172},
	}
	for _, tt := range tests {
		gotHost, gotPort := parseServerFlag(tt.in)
		if gotHost != tt.wantHost || gotPort != tt.wantPort {
			t.Errorf("parseServerFlag(%q) = (%q, %d); want (%q, %d)", tt.in, gotHost, gotPort, tt.wantHost, tt.wantPort)
		}
	}
}

func TestMaskDSN(t *testing.T) {
	tests := []struct {
		driver string
		dsn    string
		want   string
	}{
		{"sqlite", "/path/to/my.db", "sqlite:my.db"},
		{"sqlite", "", "sqlite:(path)"},
		{"postgres", "postgres://user:pass@localhost:5432/db", "postgres://user:***@localhost:5432/db"},
		{"postgres", "host=localhost password=secret user=me", "host=localhost password=*** user=me"},
		{"mysql", "user:pass@tcp(localhost:3306)/db?secret=123", "user:***@tcp(localhost:3306)/db?secret=***"},
	}
	for _, tt := range tests {
		if got := maskDSN(tt.driver, tt.dsn); got != tt.want {
			t.Errorf("maskDSN(%q, %q) = %q; want %q", tt.driver, tt.dsn, got, tt.want)
		}
	}
}

func TestSqlOpAndPreview(t *testing.T) {
	sql := "  SELECT * FROM users\nWHERE id = 1  "
	op, preview := sqlOpAndPreview(sql)
	if op != "select" {
		t.Errorf("op = %q; want %q", op, "select")
	}
	if preview != "SELECT * FROM users WHERE id = 1" {
		t.Errorf("preview = %q; want %q", preview, "SELECT * FROM users WHERE id = 1")
	}
}

func TestArgTypes(t *testing.T) {
	var b []byte
	args := []any{1, "string", nil, &b}
	got := argTypes(args)
	want := []string{"int", "string", "null", "[]byte"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("argTypes() = %v; want %v", got, want)
	}
}
