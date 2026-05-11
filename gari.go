package gari

import (
	"database/sql"
	"errors"
	"log/slog"
	"os"

	"github.com/goark/errs"
)

type (
	// Gari configuration.
	Gari struct {
		Logger *slog.Logger                  // Default logger.
		Debug  func(msg string, args ...any) // Debug level log func.
		Info   func(msg string, args ...any) // Info level log func.
		Warn   func(msg string, args ...any) // Warn level log func.
		Error  func(msg string, args ...any) // Error level log func.

		BeforeInsert func(r *Record)
		AfterInsert  func(r *Record)
		BeforeUpdate func(r *Record)
		AfterUpdate  func(r *Record)
		BeforeDelete func(r *Record)
		AfterDelete  func(r *Record)

		db          *sql.DB    // Pointer to database pool.
		isClosed    bool       // Flag to see if db is closed.
		dialect     Dialect    // Dialects of DB.
		stringQuote string     // Quotation of strings.
		idQuote     string     // Quotation of identifiers.
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
func Open(db *sql.DB, dialect Dialect, fns ...OptionFunc) (*Gari, error) {
	g := Gari{
		db:          db,
		isClosed:    false,
		dialect:     dialect,
		stringQuote: `'`,
		idQuote:     `"`,
		tables:      make([]*Table, 0),
	}
	// Setup logger.
	logOpts := slog.HandlerOptions{Level: slog.LevelDebug}
	g.Logger = slog.New(slog.NewJSONHandler(os.Stderr, &logOpts))
	g.Debug = g.Logger.Debug
	g.Info = g.Logger.Info
	g.Warn = g.Logger.Warn
	g.Error = g.Logger.Error
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
		g.infoLog("sql.DB.Close()")
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
	return &b
}

// Execute SELECT SQL statement.
func (g *Gari) Select(query string) *selectExecutor {
	return newSelectExecutor(g, query)
}

// Begin transaction.
func (g *Gari) BeginTx() *txBuilder {
	g.txBuilder = newTxBuilder(g)
	return g.txBuilder
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
