package gari

type (
	SqliteDialect struct {
	}
)

func (d SqliteDialect) QuerySuffix() string {
	return ";"
}

func (d SqliteDialect) BindVar(i int) string {
	return "?"
}
