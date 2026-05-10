package gari

import (
	"io"
)

type (
	SqliteDialect struct {
	}
)

// Single quote.
func (d SqliteDialect) QuoteString(w io.StringWriter, str string) {
	w.WriteString(`'`)
	w.WriteString(str)
	w.WriteString(`'`)
}

// Double quote.
func (d SqliteDialect) QuoteTable(w io.StringWriter, table string) {
	w.WriteString(`"`)
	w.WriteString(table)
	w.WriteString(`"`)
}

// Double quote.
func (d SqliteDialect) QuoteColumn(w io.StringWriter, column string) {
	w.WriteString(`"`)
	w.WriteString(column)
	w.WriteString(`"`)
}

// Double quote.
func (d SqliteDialect) QuoteAlias(w io.StringWriter, alias string) {
	w.WriteString(`"`)
	w.WriteString(alias)
	w.WriteString(`"`)
}

func (d SqliteDialect) QuerySuffix(w io.StringWriter) {
	w.WriteString(`;`)
}

func (d SqliteDialect) BindVar(w io.StringWriter, i int) {
	w.WriteString(`?`)
}
