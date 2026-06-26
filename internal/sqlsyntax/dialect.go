package sqlsyntax

type Dialect string

const (
	Generic  Dialect = "generic"
	Postgres Dialect = "postgres"
	MySQL    Dialect = "mysql"
	SQLite   Dialect = "sqlite"
	TSQL     Dialect = "tsql" // SQL Server
)
