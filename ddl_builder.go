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
	dialect := b.table.gari.dialect
	var sb strings.Builder
	sb.WriteString(`CREATE TABLE `)
	sb.WriteString(`IF NOT EXISTS `)
	sb.WriteString(b.table.name)
	sb.WriteString(` (`)
	for i, col := range b.table.columns {
		if i > 0 {
			sb.WriteString(`, `)
		}
		// Add column DDL token.
		sb.WriteString(col.query())
	}
	sb.WriteString(`)`)
	dialect.QuerySuffix(&sb)
	return sb.String(), nil
}
