package gari

import (
	"errors"
	"log/slog"
	"strings"

	"github.com/goark/errs"
)

type (
	// Builder to build DDL SQL.
	ddlBuilder struct {
		table *Table // Pointer to Table.
	}
)

var (
	ErrDDLBuild = errors.New("gari.ErrDDLBuild")
)

// Create new DDL builder instance.
func newDdlBuilder(t *Table) *ddlBuilder {
	db := ddlBuilder{
		table: t,
	}
	return &db
}

// Build DDL SQL.
func (b *ddlBuilder) build() (string, error) {
	tokens := make([]string, 0)
	tokens = append(tokens, "CREATE", "TABLE")
	tokens = append(tokens, "IF", "NOT", "EXISTS")
	tokens = append(tokens, b.table.name+"(")
	for i, col := range b.table.columns {
		// Add column DDL token.
		token, err := col.ddl()
		if err != nil {
			err = errs.Wrap(ErrDDLBuild, errs.WithCause(err),
				errs.WithContext("table", b.table.name),
				errs.WithContext("column", col.String()))
			b.gari().errorLog(err.Error())
			return "", err
		}
		// Add comma.
		if i < (len(b.table.columns) - 1) {
			token += ","
		}
		tokens = append(tokens, token)
	}
	tokens = append(tokens, ")")
	ddl := strings.Join(tokens, " ") + ";"
	b.gari().infoLog("ddlBuilder.build()",
		slog.String("SQL", ddl))
	return ddl, nil
}

func (b *ddlBuilder) gari() *Gari {
	return b.table.gari
}
