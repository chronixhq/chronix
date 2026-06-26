package sqlsyntax

import "testing"

func TestValidateSQL(t *testing.T) {
	type args struct {
		d         Dialect
		sql       string
		variables map[string]any
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// Empty and whitespace-only should be considered valid (no-op)
		{
			name: "Empty string is valid",
			args: args{d: Postgres, sql: ""},
			want: true,
		},
		{
			name: "Whitespace only is valid",
			args: args{d: SQLite, sql: "  \n\t  "},
			want: true,
		},
		// Postgres
		{
			name: "Postgres: simple SELECT valid",
			args: args{d: Postgres, sql: "SELECT 3;"},
			want: true,
		},
		{
			name: "Postgres: invalid keyword",
			args: args{d: Postgres, sql: "SELEC 1"},
			want: false,
		},
		{
			name: "Postgres: multiple statements valid",
			args: args{d: Postgres, sql: "SELECT 1; SELECT 2;"},
			want: true,
		},
		// MySQL
		{
			name: "MySQL: simple SELECT valid",
			args: args{d: MySQL, sql: "SELECT 1;"},
			want: true,
		},
		{
			name: "MySQL: invalid token",
			args: args{d: MySQL, sql: "SELECTE 1"},
			want: false,
		},
		{
			name: "MySQL: multiple statements valid",
			args: args{d: MySQL, sql: "SELECT 1; SELECT 2;"},
			want: true,
		},
		// SQLite
		{
			name: "SQLite: create table valid",
			args: args{d: SQLite, sql: "CREATE TABLE t (id INTEGER)"},
			want: true,
		},
		{
			name: "SQLite: malformed CREATE invalid",
			args: args{d: SQLite, sql: "CREATE TABL t (id INTEGER)"},
			want: false,
		},
		// T-SQL
		{
			name: "TSQL: SELECT TOP valid",
			args: args{d: TSQL, sql: "SELECT TOP 1 1"},
			want: true,
		},
		{
			name: "TSQL: invalid TOP keyword",
			args: args{d: TSQL, sql: "SELECT TOPX 1 1"},
			want: false,
		},
		// Template variable tests
		{
			name: "Template vars: provided -> valid (Postgres)",
			args: args{d: Postgres, sql: "SELECT {{n}};", variables: map[string]any{"n": 1}},
			want: true,
		},
		{
			name: "Template vars: missing -> invalid (Postgres)",
			args: args{d: Postgres, sql: "SELECT {{n}};"},
			want: false,
		},
		{
			name: "Template vars: multiple occurrences -> valid (Postgres)",
			args: args{d: Postgres, sql: "SELECT {{foo}} + {{foo}} + {{bar}};", variables: map[string]any{"foo": 2, "bar": 3}},
			want: true,
		},
		{
			name: "Template vars: provided -> valid (MySQL)",
			args: args{d: MySQL, sql: "SELECT {{n}};", variables: map[string]any{"n": 7}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, errStr := ValidateSQL(tt.args.d, tt.args.sql, tt.args.variables)
			if got != tt.want {
				t.Errorf("ValidateSQL() got = %v, want %v, err %s (dialect=%s, sql=%q)", got, tt.want, errStr, tt.args.d, tt.args.sql)
			}
		})
	}
}
