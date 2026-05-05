package gari

import (
	"database/sql"
	"errors"

	"github.com/goark/errs"
)

type (
	// Gari configuration.
	Gari struct {
		isClosed    bool     // Flag to see if db is closed.
		db          *sql.DB  // Pointer to database pool.
		stringQuote string   // Quotation of strings.
		columnQuote string   // Quotation of columns.
		tables      []*Table // Pointer to tables.
	}
	// Option func.
	OptionFunc func(*Gari) error
)

var (
	ErrOption = errors.New("gari.ErrOption")
)

// Open DB connection.
func Open(db *sql.DB, fns ...OptionFunc) (*Gari, error) {
	g := Gari{
		isClosed:    false,
		db:          db,
		stringQuote: `'`,
		columnQuote: `"`,
		tables:      make([]*Table, 0),
	}
	// Parse options.
	for _, fn := range fns {
		if err := fn(&g); err != nil {
			err = errs.Wrap(ErrOption, errs.WithCause(err))
			return nil, err
		}
	}
	return &g, nil
}

// Close connection.
func (g *Gari) Close() error {
	if g.isClosed {
		return nil
	}
	// Close database connection.
	defer func() {
		g.db.Close()
		g.isClosed = true
	}()
	// Close tables.
	for _, table := range g.tables {
		table.closeTable()
	}
	return nil
}

// Start table builder sequence.
func (g *Gari) Table(name string) *TableBuilder {
	b := TableBuilder{
		table: newTable(g, name),
	}
	g.tables = append(g.tables, b.table)
	// Add id column.
	b.AddInt64Column("id", PrimaryKey(true), NotNull())
	// Add timestamp columns.
	b.AddTimeColumn("created_at", NotNull())
	b.AddTimeColumn("updated_at", NotNull())
	b.AddTimeColumn("deleted_at", DefaultNull())
	return &b
}
