package gari

import (
	"context"
	"errors"
	"strings"

	"github.com/goark/errs"
)

// Executor to migrate table with DDL SQL.
type migrator struct {
	table      *Table // Pointer to Table.
	queryCache string // Cache of query.
}

var (
	ErrMigrateExec = errors.New("gari.ErrMigrateExec")
)

// Create new DDL builder instance.
func newMigrator(t *Table) *migrator {
	db := migrator{
		table: t,
	}
	return &db
}

// Build DDL SQL.
func (m *migrator) query() string {
	if len(m.queryCache) > 0 {
		return m.queryCache
	}
	dialect := m.gari().dialect
	var sb strings.Builder
	sb.WriteString(`CREATE TABLE `)
	sb.WriteString(`IF NOT EXISTS `)
	sb.WriteString(m.table.Name)
	sb.WriteString(` (`)
	for i, col := range m.table.columns {
		if i > 0 {
			sb.WriteString(`, `)
		}
		// Add column DDL SQL query.
		sb.WriteString(col.query())
	}
	sb.WriteString(`)`)
	dialect.QuerySuffix(&sb)
	m.queryCache = sb.String()
	return m.queryCache
}

// Execute DDL SQL query.
func (m *migrator) exec(ctx context.Context) error {
	_, _, err := m.gari().Exec(ctx, m.query())
	if err != nil {
		return errs.Wrap(ErrMigrateExec, errs.WithCause(err))
	}
	return nil
}

func (m *migrator) gari() *Gari {
	return m.table.gari
}
