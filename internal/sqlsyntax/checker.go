package sqlsyntax

import (
	"chronix/pkg/sqlutil"
	"fmt"
	"regexp"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	mysql "chronix/internal/sqlsyntax/gen/mysql"
	pg "chronix/internal/sqlsyntax/gen/postgres"
	sqlite "chronix/internal/sqlsyntax/gen/sqlite"
	tsql "chronix/internal/sqlsyntax/gen/tsql"
)

// ValidateSQL returns (true, "") if the SQL is syntactically valid for the given dialect.
// Otherwise it returns (false, "short reason").
func ValidateSQL(d Dialect, sql string, variables map[string]any) (bool, string) {
	sql = strings.TrimSpace(sql)
	if sql == "" {
		return true, ""
	}

	// Handle template variables of the form {{varname}}
	re := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*}}`)
	matches := re.FindAllStringSubmatch(sql, -1)
	if len(matches) > 0 {
		seen := make(map[string]struct{})
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			name := m[1]
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			if _, ok := variables[name]; !ok {
				return false, fmt.Sprintf("template variable %q is defined in SQL but not provided", name)
			}
		}
		sql = sqlutil.ResolveSQLTemplateVariables(sql, variables)
	}

	switch d {
	case Generic:
		fallthrough
	case Postgres:
		if err := checkPostgres(sql); err != nil {
			return false, err.Error()
		}
	case MySQL:
		if err := checkMySQL(sql); err != nil {
			return false, err.Error()
		}
	case SQLite:
		if err := checkSQLite(sql); err != nil {
			return false, err.Error()
		}
	case TSQL:
		if err := checkTSQL(sql); err != nil {
			return false, err.Error()
		}
	default:
		return false, "unsupported dialect"
	}
	return true, ""
}

func checkPostgres(sql string) error {
	is := antlr.NewInputStream(sql)
	lex := pg.NewPostgreSQLLexer(is)

	el := newErrorListener()
	lex.RemoveErrorListeners()
	lex.AddErrorListener(el)

	tokens := antlr.NewCommonTokenStream(lex, antlr.TokenDefaultChannel)
	p := pg.NewPostgreSQLParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(el)

	// Entry rule in grammars-v4 PostgreSQL is `Parse`
	_ = p.Root()

	return el.Err()
}

func checkMySQL(sql string) error {
	is := antlr.NewInputStream(sql)
	lex := mysql.NewMySQLLexer(is)

	el := newErrorListener()
	lex.RemoveErrorListeners()
	lex.AddErrorListener(el)

	tokens := antlr.NewCommonTokenStream(lex, antlr.TokenDefaultChannel)
	p := mysql.NewMySQLParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(el)

	// Entry rule in grammars-v4 MySQL is typically `SqlStatements`
	_ = p.Query()

	return el.Err()
}

func checkSQLite(sql string) error {
	is := antlr.NewInputStream(sql)
	lex := sqlite.NewSQLiteLexer(is)

	el := newErrorListener()
	lex.RemoveErrorListeners()
	lex.AddErrorListener(el)

	tokens := antlr.NewCommonTokenStream(lex, antlr.TokenDefaultChannel)
	p := sqlite.NewSQLiteParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(el)

	// Entry rule in grammars-v4 SQLite is `Parse`
	_ = p.Parse()

	return el.Err()
}

func checkTSQL(sql string) error {
	is := antlr.NewInputStream(sql)
	lex := tsql.NewTSqlLexer(is)

	el := newErrorListener()
	lex.RemoveErrorListeners()
	lex.AddErrorListener(el)

	tokens := antlr.NewCommonTokenStream(lex, antlr.TokenDefaultChannel)
	p := tsql.NewTSqlParser(tokens)
	p.RemoveErrorListeners()
	p.AddErrorListener(el)

	// Entry rule in grammars-v4 T-SQL is `Tsql_file`
	_ = p.Tsql_file()

	return el.Err()
}
