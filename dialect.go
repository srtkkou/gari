package gari

import (
	"io"
)

type (
	Dialect interface {
		QuoteString(w io.StringWriter, str string)
		QuoteTable(w io.StringWriter, table string)
		QuoteColumn(w io.StringWriter, column string)
		QuoteAlias(w io.StringWriter, alias string)
		QuerySuffix(w io.StringWriter)
		BindVar(w io.StringWriter, i int)
	}
)
