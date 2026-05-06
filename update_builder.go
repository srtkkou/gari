package gari

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build UPDATE SQL.
	updateBuilder struct {
		table    *Table    // Pointer to table.
		columns  []*column // Slice of column pointers.
		query    string    // UPDATE SQL query.
		prepared *sql.Stmt // Prepared statement pointer.
		records  []*Record // Records to update.
		err      error     // Error.

		debugLog func(string, ...any) // Shorthand for Gari.debugLog().
		infoLog  func(string, ...any) // Shorthand for Gari.infoLog().
		warnLog  func(string, ...any) // Shorthand for Gari.warnLog().
		errorLog func(string, ...any) // Shorthand for Gari.errorLog().
	}
)

var (
	ErrUpdatePrepare      = errors.New("gari.ErrUpdatePrepare")
	ErrUpdateValuesEncode = errors.New("gari.ErrUpdateValuesEncode")
	ErrUpdateValuesDecode = errors.New("gari.ErrUpdateValuesDecode")
	ErrUpdateRecordsEmpty = errors.New("gari.ErrUpdateRecordsEmpty")
	ErrUpdateExec         = errors.New("gari.ErrUpdateExec")
	ErrUpdateRowsAffected = errors.New("gari.ErrUpdateRowsAffected")
	ErrUpdateRowCount     = errors.New("gari.ErrUpdateRowCount")
)

// Create new updateBuilder instance.
func newUpdateBuilder(t *Table) *updateBuilder {
	b := updateBuilder{
		table:    t,
		columns:  make([]*column, 0, len(t.columnNames)),
		records:  make([]*Record, 0),
		debugLog: t.gari.debugLog,
		infoLog:  t.gari.infoLog,
		warnLog:  t.gari.warnLog,
		errorLog: t.gari.errorLog,
	}
	// Set columns except primary key column.
	for _, name := range t.columnNames {
		col := t.columns[name]
		if !col.primary {
			b.columns = append(b.columns, col)
		}
	}
	// Build UPDATE SQL query.
	b.query = b.buildQuery()
	b.infoLog("newUpdateBuilder",
		slog.String("query", b.query))
	return &b
}

// Add values to update.
func (b *updateBuilder) Values(ptrs ...any) *updateBuilder {
	if b.err != nil {
		return b
	}
	// Copy columns and add primary key column to end.
	columns := make([]*column, 0, len(b.columns)+1)
	columns = append(columns, b.columns...)
	// TODO: Fix to use non-id named pkey column.
	pkeyCol := b.table.columns["id"]
	columns = append(columns, pkeyCol)
	for _, ptr := range ptrs {
		// Convert to msgpack.
		blob, err := msgpack.Marshal(ptr)
		if err != nil {
			b.err = errs.Wrap(ErrUpdateValuesEncode, errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			b.errorLog(b.err.Error())
			return b
		}
		// Unmarshal msgpack to map[string]encoded.
		m, err := splitToMap(blob)
		if err != nil {
			b.err = errs.Wrap(ErrUpdateValuesDecode, errs.WithCause(err),
				errs.WithContext("ptr", ptr),
				errs.WithContext("msgpack", blob))
			return b
		}
		// Build record.
		cells := make([]*cell, len(columns))
		for i, col := range columns {
			cells[i] = newCell(col)
			cells[i].blob = m[col.fieldName]
		}
		r := newRecord(cells)
		b.records = append(b.records, r)
	}
	return b
}

// Execute UPDATE SQL statement.
func (b *updateBuilder) Exec(ctx context.Context) error {
	defer func() {
		b.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if b.err != nil {
		return b.err
	}
	// Check if records are not empty.
	if len(b.records) == 0 {
		b.err = errs.Wrap(ErrUpdateRecordsEmpty,
			errs.WithContext("query", b.query))
		b.errorLog(b.err.Error())
		return b.err
	}
	for _, r := range b.records {
		// TODO:BEFORE UPDATE
		// Execute prepared statement.
		args := r.args()
		startedAt := time.Now()
		result, err := b.prepared.ExecContext(ctx, args...)
		if err != nil {
			b.err = errs.Wrap(ErrUpdateExec, errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			b.errorLog(b.err.Error())
			return b.err
		}
		b.infoLog("Execute UPDATE SQL.",
			slog.String("query", b.query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			b.err = errs.Wrap(ErrUpdateRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args))
			b.errorLog(b.err.Error())
			return b.err
		}
		if count != 1 {
			b.err = errs.Wrap(ErrUpdateRowCount,
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			b.errorLog(b.err.Error())
			return b.err
		}
		// TODO:AFTER UPDATE
	}
	return nil
}

// Close prepared statement.
func (b *updateBuilder) closePreparedStmt() {
	if b.prepared != nil {
		b.prepared.Close()
	}
}

// Build SQL statement.
func (b *updateBuilder) buildQuery() string {
	tokens := []string{"UPDATE"}
	// Add table name.
	quote := b.table.gari.columnQuote
	table := fmt.Sprintf("%s%s%s", quote, b.table.name, quote)
	tokens = append(tokens, table, "SET")
	// Add column names and placeholders.
	count := 0
	for _, col := range b.columns {
		ph := b.table.gari.placeholder(count)
		if count < (len(b.columns) - 1) {
			ph += ","
		}
		tokens = append(tokens, col.name, "=", ph)
		count++
	}
	// Add WHERE statement.
	tokens = append(tokens, "WHERE")
	// TODO: Change needed to use non ID column name.
	pkey := fmt.Sprintf("%s%s%s", quote, "id", quote)
	tokens = append(tokens, pkey, "=")
	ph := b.table.gari.placeholder(count) + ";"
	tokens = append(tokens, ph)
	return strings.Join(tokens, " ")
}

// Prepare UPDATE SQL statement.
func (b *updateBuilder) prepareStmt() {
	startedAt := time.Now()
	var err error
	db := b.table.gari.db
	b.prepared, err = db.Prepare(b.query)
	if err != nil {
		b.err = errs.Wrap(ErrUpdatePrepare, errs.WithCause(err),
			errs.WithContext("query", b.query),
			errs.WithContext("duration", time.Since(startedAt)))
		b.errorLog(b.err.Error())
	}
	b.infoLog("Prepare UPDATE SQL.",
		slog.String("query", b.query),
		slog.Duration("duration", time.Since(startedAt)))
}
