package gari

import (
	"database/sql"
	"errors"

	"github.com/goark/errs"
)

type (
	// Gari configuration.
	Gari struct {
		db          *sql.DB // Pointer to database pool.
		stringQuote string  // Quotation of strings.
		columnQuote string  // Quotation of columns.
	}
	// Option func.
	OptionFunc func(*Gari) error
)

var (
	ErrOption = errors.New("gari.ErrOption")
	ErrClose  = errors.New("gari.ErrClose")
)

// Initialize.
func New(db *sql.DB, fns ...OptionFunc) (*Gari, error) {
	g := Gari{
		db:          db,
		stringQuote: `'`,
		columnQuote: `"`,
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
	if err := g.db.Close(); err != nil {
		return errs.Wrap(ErrClose, errs.WithCause(err))
	}
	return nil
}

// Start table builder sequence.
func (g *Gari) Table(name string) *TableBuilder {
	b := TableBuilder{
		table: newTable(g, name),
	}
	// Add id column.
	b.AddInt64Column("id", func(c *Column) {
		c.primary = true
		c.autoIncrement = true
		c.notNull = true
	})
	// Add timestamp columns.
	b.AddTimeColumn("created_at", NotNull())
	b.AddTimeColumn("updated_at", NotNull())
	b.AddTimeColumn("deleted_at", DefaultNull())
	return &b
}
