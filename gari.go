package gari

import (
	"database/sql"
	"errors"

	"github.com/goark/errs"
)

type (
	// Gari configuration.
	Gari struct {
		Debug func(msg string, args ...any) // Debug level log func.
		Info  func(msg string, args ...any) // Info level log func.
		Warn  func(msg string, args ...any) // Warn level log func.
		Error func(msg string, args ...any) // Error level log func.

		db          *sql.DB    // Pointer to database pool.
		isClosed    bool       // Flag to see if db is closed.
		stringQuote string     // Quotation of strings.
		columnQuote string     // Quotation of columns.
		tables      []*Table   // Pointer to tables.
		txBuilder   *txBuilder // Pointer to txBuilder.
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
	// Close transaction.
	if g.txBuilder != nil {
		g.txBuilder.Rollback()
	}
	return nil
}

// Start table builder sequence.
func (g *Gari) Table(name string) *TableBuilder {
	b := TableBuilder{
		table: newTable(g, name),
	}
	g.tables = append(g.tables, b.table)
	// TODO: Remove ID/TIMESTAMP columns below.
	// Add id column.
	b.AddInt64Column("id", PrimaryKey(true), NotNull())
	// Add timestamp columns.
	b.AddTimeColumn("created_at", NotNull())
	b.AddTimeColumn("updated_at", NotNull())
	b.AddTimeColumn("deleted_at", DefaultNull())
	return &b
}

// Begin transaction.
func (g *Gari) BeginTx() *txBuilder {
	g.txBuilder = newTxBuilder(g)
	return g.txBuilder
}

// Get placeholder.
// TODO: It should return $1,$2... for postgres.
func (g *Gari) placeholder(i int) string {
	return "?"
}

// Output debug log.
//
//lint:ignore U1000 Defined for future use.
func (g *Gari) debugLog(msg string, args ...any) {
	if g.Debug != nil {
		g.Debug(msg, args...)
	}
}

// Output info log.
//
//lint:ignore U1000 Defined for future use.
func (g *Gari) infoLog(msg string, args ...any) {
	if g.Info != nil {
		g.Info(msg, args...)
	}
}

// Output warn log.
//
//lint:ignore U1000 Defined for future use.
func (g *Gari) warnLog(msg string, args ...any) {
	if g.Warn != nil {
		g.Warn(msg, args...)
	}
}

// Output error log.
//
//lint:ignore U1000 Defined for future use.
func (g *Gari) errorLog(msg string, args ...any) {
	if g.Error != nil {
		g.Error(msg, args...)
	}
}
