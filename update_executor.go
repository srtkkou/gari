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
	// Executor to run UPDATE SQL.
	updateExecutor struct {
		table    *Table    // Pointer to table.
		columns  []*column // Slice of column pointers.
		query    string    // UPDATE SQL query.
		prepared *sql.Stmt // Prepared statement pointer.
		records  []*Record // Records to update.
		err      error     // Error.
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

// Create new updateExecutor instance.
func newUpdateExecutor(t *Table) *updateExecutor {
	b := updateExecutor{
		table:   t,
		columns: make([]*column, 0, (len(t.columns) - 1)),
		records: make([]*Record, 0),
	}
	// Set columns except primary key column.
	for _, col := range t.columns {
		if col.primary {
			continue
		}
		b.columns = append(b.columns, col)
	}
	// Build UPDATE SQL query.
	b.query = b.buildQuery()
	b.gari().infoLog("newUpdateExecutor",
		slog.String("query", b.query))
	return &b
}

// Add values to update.
func (b *updateExecutor) Values(ptrs ...any) *updateExecutor {
	if b.err != nil {
		return b
	}
	// Copy columns and add primary key column to end.
	columns := make([]*column, 0, len(b.columns)+1)
	columns = append(columns, b.columns...)
	// TODO: Fix to use non-id named pkey column.
	pkeyCol := b.table.column("id")
	columns = append(columns, pkeyCol)
	for _, ptr := range ptrs {
		// Convert to msgpack.
		blob, err := msgpack.Marshal(ptr)
		if err != nil {
			b.err = errs.Wrap(ErrUpdateValuesEncode, errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			b.gari().errorLog(b.err.Error())
			return b
		}
		// Unmarshal msgpack to map[string]encoded.
		m, err := splitToMap(blob)
		if err != nil {
			b.err = errs.Wrap(ErrUpdateValuesDecode, errs.WithCause(err),
				errs.WithContext("ptr", ptr),
				errs.WithContext("msgpack", blob))
			b.gari().errorLog(b.err.Error())
			return b
		}
		// Build record.
		values := make([]*value, len(columns))
		for i, col := range columns {
			values[i] = newValue(col)
			values[i].blob = m[col.fieldName]
		}
		r := newRecord(values)
		b.records = append(b.records, r)
	}
	return b
}

// Execute UPDATE SQL statement.
func (b *updateExecutor) Exec(ctx context.Context) error {
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
		b.gari().errorLog(b.err.Error())
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
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		b.gari().infoLog("Execute UPDATE SQL.",
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
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		if count != 1 {
			b.err = errs.Wrap(ErrUpdateRowCount,
				errs.WithContext("query", b.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			b.gari().errorLog(b.err.Error())
			return b.err
		}
		// TODO:AFTER UPDATE
	}
	return nil
}

// Close prepared statement.
func (b *updateExecutor) closePreparedStmt() {
	if b.prepared != nil {
		b.prepared.Close()
	}
}

// Build SQL statement.
func (b *updateExecutor) buildQuery() string {
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
func (b *updateExecutor) prepareStmt() {
	startedAt := time.Now()
	var err error
	db := b.table.gari.db
	b.prepared, err = db.Prepare(b.query)
	if err != nil {
		b.err = errs.Wrap(ErrUpdatePrepare, errs.WithCause(err),
			errs.WithContext("query", b.query),
			errs.WithContext("duration", time.Since(startedAt)))
		b.gari().errorLog(b.err.Error())
	}
	b.gari().infoLog("Prepare UPDATE SQL.",
		slog.String("query", b.query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (b *updateExecutor) gari() *Gari {
	return b.table.gari
}
