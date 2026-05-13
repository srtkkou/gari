package gari

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"time"

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

		db          *sql.DB // Pointer to database pool.
		isClosed    bool    // Flag to see if db is closed.
		dialect     Dialect // Dialects of DB.
		stringQuote string  // Quotation of strings.
		idQuote     string  // Quotation of identifiers.

		tables   []*Table          // Slice of table ptrs.
		tableMap map[string]*Table // Map of table ptrs.

		txBuilder *txBuilder // Pointer to txBuilder.
	}
	// Option func.
	OptionFunc func(*Gari) error
)

var (
	ErrOption          = errors.New("gari.ErrOption")
	ErrSqlExec         = errors.New("gari.ErrSqlExec")
	ErrLastInsertId    = errors.New("gari.ErrLastInsertId")
	ErrRowsAffected    = errors.New("gari.ErrRowsAffected")
	ErrSqlPrepare      = errors.New("gari.ErrSqlPrepare")
	ErrSqlExecPrepared = errors.New("gari.ErrSqlExecPrepared")
)

// Open DB connection.
func Open(db *sql.DB, dialect Dialect, fns ...OptionFunc) (*Gari, error) {
	g := &Gari{
		db:          db,
		isClosed:    false,
		dialect:     dialect,
		stringQuote: `'`,
		idQuote:     `"`,
		tables:      make([]*Table, 0),
		tableMap:    make(map[string]*Table, 0),
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
		if err := fn(g); err != nil {
			err = errs.Wrap(ErrOption, errs.WithCause(err))
			return nil, err
		}
	}
	return g, nil
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
	table := newTable(g, name)
	b := TableBuilder{table: table}
	g.addTable(table)
	return &b
}

// Start schema builder sequence.
func (g *Gari) TableWithSchema(
	name, schemaName string,
) *TableBuilder {
	table := newTable(g, name)
	table.SchemaName = schemaName
	b := TableBuilder{table: table}
	g.addTable(table)
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

func (g *Gari) Exec(
	ctx context.Context, query string, args ...any,
) (int64, int64, error) {
	startedAt := time.Now()
	// Execute query.
	result, err := g.db.ExecContext(ctx, query, args...)
	duration := time.Since(startedAt)
	if err != nil {
		err = errs.Wrap(ErrSqlExec, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("args", args),
			errs.WithContext("duration", duration))
		g.errorLog(err.Error())
		return 0, 0, err
	}
	id, count, err := g.parseSqlResult(result)
	if err != nil {
		return 0, 0, err
	}
	g.infoLog("gari.Exec",
		slog.String("query", query),
		slog.Any("args", args),
		slog.Duration("duration", duration),
		slog.Int64("lastInsertId", id),
		slog.Int64("rowsAffected", count))
	return id, count, nil
}

func (g *Gari) prepare(query string) (*sql.Stmt, error) {
	startedAt := time.Now()
	// Prepare query.
	stmt, err := g.db.Prepare(query)
	duration := time.Since(startedAt)
	if err != nil {
		err = errs.Wrap(ErrSqlPrepare, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", duration))
		g.errorLog(err.Error())
		return nil, err
	}
	g.infoLog("gari.prepare",
		slog.String("query", query),
		slog.Duration("duration", duration))
	return stmt, nil
}

func (g *Gari) execPrepared(
	ctx context.Context, stmt *sql.Stmt, args ...any,
) (int64, int64, error) {
	startedAt := time.Now()
	// Execute query.
	result, err := stmt.ExecContext(ctx, args...)
	duration := time.Since(startedAt)
	if err != nil {
		err = errs.Wrap(ErrSqlExecPrepared, errs.WithCause(err),
			errs.WithContext("stmt", stmt),
			errs.WithContext("args", args),
			errs.WithContext("duration", duration))
		g.errorLog(err.Error())
		return 0, 0, err
	}
	id, count, err := g.parseSqlResult(result)
	if err != nil {
		return 0, 0, err
	}
	g.infoLog("gari.execPrepared",
		slog.Any("stmt", stmt),
		slog.Any("args", args),
		slog.Duration("duration", duration),
		slog.Int64("lastInsertId", id),
		slog.Int64("rowsAffected", count))
	return id, count, nil
}

func (g *Gari) parseSqlResult(result sql.Result) (int64, int64, error) {
	// Parse last insert id.
	id, err := result.LastInsertId()
	if err != nil {
		err = errs.Wrap(ErrLastInsertId, errs.WithCause(err),
			errs.WithContext("result", result))
		g.errorLog(err.Error())
		return 0, 0, err
	}
	// Parse rows affected.
	count, err := result.RowsAffected()
	if err != nil {
		err = errs.Wrap(ErrLastInsertId, errs.WithCause(err),
			errs.WithContext("result", result))
		g.errorLog(err.Error())
		return 0, 0, err
	}
	return id, count, nil
}

func (g *Gari) addTable(table *Table) {
	g.tables = append(g.tables, table)
	g.tableMap[table.Name] = table
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
