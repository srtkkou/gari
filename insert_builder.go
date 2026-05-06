package gari

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build INSERT SQL.
	insertBuilder struct {
		table    *Table    // Pointer to table.
		columns  []*column // Slice of column pointers.
		query    string    // INSERT SQL query.
		prepared *sql.Stmt // Prepared statement pointer.
		records  []*Record // Records to insert.
		err      error     // Error.

		debugLog func(string, ...any) // Shorthand for Gari.debugLog().
		infoLog  func(string, ...any) // Shorthand for Gari.infoLog().
		warnLog  func(string, ...any) // Shorthand for Gari.warnLog().
		errorLog func(string, ...any) // Shorthand for Gari.errorLog().
	}
)

var (
	ErrInsertRecordsEmpty = errors.New("gari.ErrInsertRecordsEmpty")
	ErrInsertPrepare      = errors.New("gari.ErrInsertPrepare")
	ErrInsertValuesEncode = errors.New("gari.ErrInsertValuesEncode")
	ErrInsertValuesDecode = errors.New("gari.ErrInsertValuesDecode")
	ErrInsertExec         = errors.New("gari.ErrInsertExec")
	ErrInsertRowsAffected = errors.New("gari.ErrInsertRowsAffected")
	ErrInsertRowCount     = errors.New("gari.ErrInsertRowCount")
)

// Create new insertBuilder instance.
func newInsertBuilder(t *Table) *insertBuilder {
	b := insertBuilder{
		table:    t,
		columns:  make([]*column, 0, len(t.columnNames)),
		records:  make([]*Record, 0),
		debugLog: t.gari.debugLog,
		infoLog:  t.gari.infoLog,
		warnLog:  t.gari.warnLog,
		errorLog: t.gari.errorLog,
	}
	// Set columns except primary key.
	for _, name := range t.columnNames {
		col := t.columns[name]
		if !col.primary {
			b.columns = append(b.columns, col)
		}
	}
	// Build INSERT SQL query.
	b.query = b.buildQuery()
	b.infoLog("newInsertBuilder",
		slog.String("query", b.query))
	return &b
}

// Add values to insert.
func (b *insertBuilder) Values(ptrs ...any) *insertBuilder {
	if b.err != nil {
		return b
	}
	for _, ptr := range ptrs {
		// Convert to msgpack.
		blob, err := msgpack.Marshal(ptr)
		if err != nil {
			b.err = errs.Wrap(ErrInsertValuesEncode, errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			b.errorLog(b.err.Error())
			return b
		}
		// Unmarshal msgpack to map[string]encoded.
		m, err := splitToMap(blob)
		if err != nil {
			b.err = errs.Wrap(ErrInsertValuesDecode, errs.WithCause(err),
				errs.WithContext("ptr", ptr),
				errs.WithContext("msgpack", blob))
			return b
		}
		// Build record.
		cells := make([]*cell, len(b.columns))
		for i, col := range b.columns {
			cells[i] = newCell(col)
			cells[i].blob = m[col.fieldName]
		}
		r := newRecord(cells)
		b.records = append(b.records, r)
	}
	return b
}

// Execute INSERT SQL statement.
func (b *insertBuilder) Exec(ctx context.Context) error {
	defer func() {
		b.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if b.err != nil {
		return b.err
	}
	// Check if records are not empty.
	if len(b.records) == 0 {
		b.err = errs.Wrap(ErrInsertRecordsEmpty,
			errs.WithContext("query", b.query))
		b.errorLog(b.err.Error())
		return b.err
	}
	for _, r := range b.records {
		// TODO:BEFORE UPDATE
		// Execute prepared statement.
		args := r.args()
		b.infoLog("Execute INSERT SQL.",
			slog.String("query", b.query),
			slog.Any("args", args))
		result, err := b.prepared.ExecContext(ctx, args...)
		if err != nil {
			b.err = errs.Wrap(ErrInsertExec, errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args))
			b.errorLog(b.err.Error())
			return b.err
		}
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			b.err = errs.Wrap(ErrInsertRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args))
			b.errorLog(b.err.Error())
			return b.err
		}
		if count != 1 {
			b.err = errs.Wrap(ErrInsertRowCount,
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
func (b *insertBuilder) closePreparedStmt() {
	if b.prepared != nil {
		b.prepared.Close()
	}
}

// Build INSERT SQL statement.
func (b *insertBuilder) buildQuery() string {
	tokens := []string{"INSERT", "INTO"}
	// Add table name.
	quote := b.table.gari.columnQuote
	table := fmt.Sprintf("%s%s%s", quote, b.table.name, quote)
	tokens = append(tokens, table, "(")
	// Add column names.
	for i, col := range b.columns {
		name := fmt.Sprintf("%s%s%s", quote, col.name, quote)
		if i < (len(b.columns) - 1) {
			name += ","
		}
		tokens = append(tokens, name)
	}
	// Add value placeholders.
	tokens = append(tokens, ")", "VALUES", "(")
	for i := range b.columns {
		ph := b.table.gari.placeholder(i)
		if i < (len(b.columns) - 1) {
			ph += ","
		}
		tokens = append(tokens, ph)
	}
	tokens = append(tokens, ");")
	return strings.Join(tokens, " ")
}

// Prepare INSERT SQL statement.
func (b *insertBuilder) prepareStmt() {
	b.infoLog("Prepare INSERT SQL.",
		slog.String("query", b.query))
	var err error
	db := b.table.gari.db
	b.prepared, err = db.Prepare(b.query)
	if err != nil {
		b.err = errs.Wrap(ErrInsertPrepare, errs.WithCause(err),
			errs.WithContext("query", b.query))
		b.errorLog(b.err.Error())
	}
}
