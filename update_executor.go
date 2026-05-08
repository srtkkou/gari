package gari

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
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
	ErrUpdateValuePtr     = errors.New("gari.ErrUpdateValuePtr")
	ErrUpdateNoRecords    = errors.New("gari.ErrUpdateNoRecords")
	ErrUpdateExec         = errors.New("gari.ErrUpdateExec")
	ErrUpdateRowsAffected = errors.New("gari.ErrUpdateRowsAffected")
	ErrUpdateRowCount     = errors.New("gari.ErrUpdateRowCount")
)

// Create new updateExecutor instance.
func newUpdateExecutor(t *Table) *updateExecutor {
	e := updateExecutor{
		table:   t,
		columns: make([]*column, 0, (len(t.columns) - 1)),
		records: make([]*Record, 0),
	}
	// Set columns except primary key column.
	for _, col := range t.columns {
		if col.isPrimary {
			continue
		}
		e.columns = append(e.columns, col)
	}
	// Build UPDATE SQL query.
	e.query = e.buildQuery()
	return &e
}

// Add values to update.
func (e *updateExecutor) Values(ptrs ...any) *updateExecutor {
	if e.err != nil {
		return e
	}
	// Copy columns and add primary key column to end.
	columns := make([]*column, 0, len(e.columns)+1)
	columns = append(columns, e.columns...)
	columns = append(columns, e.table.pkey)
	for _, ptr := range ptrs {
		// Build map of values from pointer to struct.
		m, err := structPtrToMap(ptr)
		if err != nil {
			e.err = errs.Wrap(ErrUpdateValuePtr,
				errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			e.gari().errorLog(e.err.Error())
			return e
		}
		// Build record.
		values := make([]*value, len(columns))
		for i, col := range columns {
			values[i] = newValueByColumn(col)
			values[i].raw = m[col.fieldName]
		}
		r := newRecord(values)
		e.records = append(e.records, r)
	}
	return e
}

// Execute UPDATE SQL statement.
func (e *updateExecutor) Exec(ctx context.Context) error {
	defer func() {
		e.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if e.err != nil {
		return e.err
	}
	// Check if records are not empty.
	if len(e.records) == 0 {
		e.err = errs.Wrap(ErrUpdateNoRecords,
			errs.WithContext("query", e.query),
			errs.WithContext("records", 0))
		e.gari().errorLog(e.err.Error())
		return e.err
	}
	for _, r := range e.records {
		// TODO:BEFORE UPDATE
		// Execute prepared statement.
		args := r.args()
		startedAt := time.Now()
		result, err := e.prepared.ExecContext(ctx, args...)
		if err != nil {
			e.err = errs.Wrap(ErrUpdateExec, errs.WithCause(err),
				errs.WithContext("query", e.query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		e.gari().infoLog("Execute UPDATE SQL.",
			slog.String("query", e.query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			e.err = errs.Wrap(ErrUpdateRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", e.query),
				errs.WithContext("args", args))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		if count != 1 {
			e.err = errs.Wrap(ErrUpdateRowCount,
				errs.WithContext("query", e.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		// TODO:AFTER UPDATE
	}
	return nil
}

// Close prepared statement.
func (e *updateExecutor) closePreparedStmt() {
	if e.prepared != nil {
		e.prepared.Close()
	}
}

// Build SQL statement.
func (e *updateExecutor) buildQuery() string {
	tokens := []string{"UPDATE"}
	// Add table name.
	tokens = append(tokens, e.table.quotedName(), "SET")
	// Add column names and placeholders.
	count := 0
	for _, col := range e.columns {
		ph := e.table.gari.placeholder(count)
		if count < (len(e.columns) - 1) {
			ph += ","
		}
		tokens = append(tokens, col.quotedName(), "=", ph)
		count++
	}
	// Add WHERE statement.
	tokens = append(tokens, "WHERE")
	pkey := e.table.pkey.quotedName()
	tokens = append(tokens, pkey, "=")
	ph := e.table.gari.placeholder(count) + ";"
	tokens = append(tokens, ph)
	return strings.Join(tokens, " ")
}

// Prepare UPDATE SQL statement.
func (e *updateExecutor) prepareStmt() {
	startedAt := time.Now()
	var err error
	e.prepared, err = e.gari().db.Prepare(e.query)
	if err != nil {
		e.err = errs.Wrap(ErrUpdatePrepare, errs.WithCause(err),
			errs.WithContext("query", e.query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari().errorLog(e.err.Error())
	}
	e.gari().infoLog("Prepare UPDATE SQL.",
		slog.String("query", e.query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (e *updateExecutor) gari() *Gari {
	return e.table.gari
}
