package gari

import (
	"errors"
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
	tokens = append(tokens, "CREATE TABLE")
	tokens = append(tokens, "IF NOT EXISTS")
	tokens = append(tokens, b.table.name+"(")
	for i, name := range b.table.columnNames {
		// Add column DDL token.
		col := b.table.columns[name]
		token, err := col.ddl()
		if err != nil {
			err = errs.Wrap(ErrDDLBuild, errs.WithCause(err),
				errs.WithContext("table", b.table.name))
			return "", err
		}
		// Add comma.
		if i < (len(b.table.columnNames) - 1) {
			token += ","
		}
		tokens = append(tokens, token)
	}
	tokens = append(tokens, ")")
	ddl := strings.Join(tokens, " ") + ";"
	return ddl, nil
}
