package gari

import (
	"strings"
)

type (
	// Builder to build DDL SQL.
	ddlBuilder struct {
		table *Table // Pointer to Table.
	}
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
	tokens = append(tokens, b.table.name, "(")
	for i, col := range b.table.columns {
		// Add column DDL token.
		token := col.query()
		// Add comma.
		if i < (len(b.table.columns) - 1) {
			token += ","
		}
		tokens = append(tokens, token)
	}
	tokens = append(tokens, ");")
	query := strings.Join(tokens, " ")
	return query, nil
}
