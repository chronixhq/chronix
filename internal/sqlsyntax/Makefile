ANTLR=java -Xmx500M -cp /usr/local/bin/antlr-4.13.1-complete.jar org.antlr.v4.Tool

gen: gen-postgres gen-mysql gen-sqlite gen-tsql

# --- PostgreSQL ---
# 1) run ANTLR from inside grammars/postgres so tokens resolve
# 2) copy Go helper files and retag package to match (-package pgparser)
gen-postgres:
	@mkdir -p gen/postgres
	@(cd grammars/postgres && \
		$(ANTLR) -Dlanguage=Go -listener -visitor \
		-package pgparser \
		-o ../../gen/postgres \
		PostgreSQLLexer.g4 PostgreSQLParser.g4)
	# bring in the helper base files (and rename package)
	@cp grammars/postgres/*.go gen/postgres/ 2>/dev/null || true
	@sed -i '' 's/^package parser$$/package pgparser/' gen/postgres/*.go 2>/dev/null || true
	@sed -i '' -E 's/^([[:space:]]*)func \((p|l) \*PostgreSQLLexer\)/\1func (this *PostgreSQLLexer)/' gen/postgres/postgresql_lexer.go || true
	@sed -i '' -E '/^func \(this \*PostgreSQLLexer\) (Action|Sempred)\(/,/^func / s/(^|[^[:alnum:]_])[lp]\./\1this./g' gen/postgres/postgresql_lexer.go || true
	@sed -i '' -E 's/(^|[^[:alnum:]_])this\./\1p./g' gen/postgres/postgresql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewPostgreSQLParser\(/,/^[[:space:]]*}/ s/^[[:space:]]*this[[:space:]]*:=[[:space:]]*new\([[:space:]]*PostgreSQLParser[[:space:]]*\)/p := new(PostgreSQLParser)/' gen/postgres/postgresql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewPostgreSQLParser\(/,/^[[:space:]]*}/ s/(ParserATNSimulator\()this,/\1p,/' gen/postgres/postgresql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewPostgreSQLParser\(/,/^[[:space:]]*}/ s/return[[:space:]]+this/return p/' gen/postgres/postgresql_parser.go || true

# --- MySQL ---
gen-mysql:
	@mkdir -p gen/mysql
	@(cd grammars/mysql && \
		$(ANTLR) -Dlanguage=Go -listener -visitor \
		-package mysqlparser \
		-o ../../gen/mysql \
		MySQLLexer.g4 MySQLParser.g4)
	# bring in the helper base files (and rename package)
	@cp grammars/mysql/*.go gen/mysql/ 2>/dev/null || true
	@sed -i '' 's/^package parser$$/package mysqlparser/' gen/mysql/*.go 2>/dev/null || true
	@sed -i '' -E 's/^([[:space:]]*)func \((p|l) \*MySQLLexer\)/\1func (this *MySQLLexer)/' gen/mysql/mysql_lexer.go || true
	@sed -i '' -E '/^func \(this \*MySQLLexer\) (Action|Sempred)\(/,/^func / s/(^|[^[:alnum:]_])[lp]\./\1this./g' gen/mysql/mysql_lexer.go || true
	@sed -i '' -E 's/(^|[^[:alnum:]_])this\./\1p./g' gen/mysql/mysql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewMySQLParser\(/,/^[[:space:]]*}/ s/^[[:space:]]*this[[:space:]]*:=[[:space:]]*new[[:space:]]*\([[:space:]]*MySQLParser[[:space:]]*\)/p := new(MySQLParser)/' gen/mysql/mysql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewMySQLParser\(/,/^[[:space:]]*}/ s/(ParserATNSimulator\()this,/\1p,/' gen/mysql/mysql_parser.go || true
	@sed -i '' -E '/^func[[:space:]]+NewMySQLParser\(/,/^[[:space:]]*}/ s/return[[:space:]]+this/return p/' gen/mysql/mysql_parser.go || true

# --- SQLite ---
gen-sqlite:
	@mkdir -p gen/sqlite
	@(cd grammars/sqlite && \
		$(ANTLR) -Dlanguage=Go -listener -visitor \
		-package sqliteparser \
		-o ../../gen/sqlite \
		SQLiteLexer.g4 SQLiteParser.g4)

gen-tsql:
	@mkdir -p gen/tsql
	@(cd grammars/tsql && \
		$(ANTLR) -Dlanguage=Go -listener -visitor \
		-package tsqlparser \
		-o ../../gen/tsql \
		TSqlLexer.g4 TSqlParser.g4)

clean:
	rm -rf gen/postgres gen/mysql gen/sqlite gen/tsql
.PHONY: gen gen-postgres gen-mysql gen-sqlite gen-tsql clean
