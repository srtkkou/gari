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
	// Builder to build INSERT SQL.
	insertBuilder struct {
		table    *Table    // Pointer to table.
		columns  []*column // Slice of columns except primary key.
		query    string    // INSERT SQL query.
		prepared *sql.Stmt // Prepared statement pointer.
		records  []*Record // Records to insert.
		err      error     // Error.
	}
)

var (
	ErrInsertPrepare      = errors.New("gari.ErrInsertPrepare")
	ErrInsertValuesEncode = errors.New("gari.ErrInsertValuesEncode")
	ErrInsertValuesDecode = errors.New("gari.ErrInsertValuesDecode")
	ErrInsertRecordsEmpty = errors.New("gari.ErrInsertRecordsEmpty")
	ErrInsertExec         = errors.New("gari.ErrInsertExec")
	ErrInsertRowsAffected = errors.New("gari.ErrInsertRowsAffected")
	ErrInsertRowCount     = errors.New("gari.ErrInsertRowCount")
)

// Create new insertBuilder instance.
func newInsertBuilder(t *Table) *insertBuilder {
	b := insertBuilder{
		table:   t,
		columns: make([]*column, 0, (len(t.columns) - 1)),
		records: make([]*Record, 0),
	}
	// Initialize columns.
	for _, col := range t.columns {
		// Skip primary key.
		if col.primary {
			continue
		}
		b.columns = append(b.columns, col)
	}
	// Build INSERT SQL query.
	b.query = b.buildQuery()
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
			b.gari().errorLog(b.err.Error())
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
		values := make([]*value, len(b.columns))
		for i, col := range b.columns {
			values[i] = newValue(col)
			values[i].blob = m[col.fieldName]
		}
		r := newRecord(values)
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
		b.gari().errorLog(b.err.Error())
		return b.err
	}
	for _, r := range b.records {
		// TODO:BEFORE INSERT
		// Execute prepared statement.
		args := r.args()
		startedAt := time.Now()
		result, err := b.prepared.ExecContext(ctx, args...)
		if err != nil {
			b.err = errs.Wrap(ErrInsertExec, errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		b.gari().infoLog("Execute INSERT SQL.",
			slog.String("query", b.query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			b.err = errs.Wrap(ErrInsertRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", b.query),
				errs.WithContext("args", args))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		if count != 1 {
			b.err = errs.Wrap(ErrInsertRowCount,
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		// TODO:AFTER INSERT
	}
	return nil
}

// Close prepared statement.
func (b *insertBuilder) closePreparedStmt() {
	if b.prepared != nil {
		b.prepared.Close()
		b.gari().infoLog("insertBuilder.closePreparedStmt()")
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
	query := strings.Join(tokens, " ")
	b.gari().debugLog("insertBuilder.buildQuery()",
		slog.String("query", query))
	return query
}

// Prepare INSERT SQL statement.
func (b *insertBuilder) prepareStmt() {
	var err error
	db := b.table.gari.db
	startedAt := time.Now()
	b.prepared, err = db.Prepare(b.query)
	if err != nil {
		b.err = errs.Wrap(ErrInsertPrepare, errs.WithCause(err),
			errs.WithContext("query", b.query),
			errs.WithContext("duration", time.Since(startedAt)))
		b.gari().errorLog(b.err.Error())
	}
	b.gari().infoLog("insertBuilder.prepareStmt()",
		slog.String("query", b.query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (b *insertBuilder) gari() *Gari {
	return b.table.gari
}
