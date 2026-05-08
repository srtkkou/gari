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
	// Executor to run INSERT SQL.
	insertExecutor struct {
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
	ErrInsertValuePtr     = errors.New("gari.ErrInsertValuePtr")
	ErrInsertNoRecords    = errors.New("gari.ErrInsertNoRecords")
	ErrInsertExec         = errors.New("gari.ErrInsertExec")
	ErrInsertRowsAffected = errors.New("gari.ErrInsertRowsAffected")
	ErrInsertRowCount     = errors.New("gari.ErrInsertRowCount")
)

// Create new insertExecutor instance.
func newInsertExecutor(t *Table) *insertExecutor {
	e := insertExecutor{
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
		e.columns = append(e.columns, col)
	}
	// Build INSERT SQL query.
	e.query = e.buildQuery()
	return &e
}

// Add values to insert.
func (e *insertExecutor) Values(ptrs ...any) *insertExecutor {
	if e.err != nil {
		return e
	}
	for _, ptr := range ptrs {
		// Build map of values from pointer to struct.
		m, err := structPtrToMap(ptr)
		if err != nil {
			e.err = errs.Wrap(ErrInsertValuePtr,
				errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			e.gari().errorLog(e.err.Error())
			return e
		}
		// Build record.
		values := make([]*value, len(e.columns))
		for i, col := range e.columns {
			values[i] = newValueByColumn(col)
			values[i].raw = m[col.fieldName]
		}
		r := newRecord(values)
		e.records = append(e.records, r)
	}
	return e
}

// Execute INSERT SQL statement.
func (e *insertExecutor) Exec(ctx context.Context) error {
	defer func() {
		e.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if e.err != nil {
		return e.err
	}
	// Check if records are not empty.
	if len(e.records) == 0 {
		e.err = errs.Wrap(ErrInsertNoRecords,
			errs.WithContext("query", e.query),
			errs.WithContext("records", 0))
		e.gari().errorLog(e.err.Error())
		return e.err
	}
	for _, r := range e.records {
		// TODO:BEFORE INSERT
		// Execute prepared statement.
		args := r.args()
		startedAt := time.Now()
		result, err := e.prepared.ExecContext(ctx, args...)
		if err != nil {
			e.err = errs.Wrap(ErrInsertExec, errs.WithCause(err),
				errs.WithContext("query", e.query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		e.gari().infoLog("Execute INSERT SQL.",
			slog.String("query", e.query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			e.err = errs.Wrap(ErrInsertRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", e.query),
				errs.WithContext("args", args))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		if count != 1 {
			e.err = errs.Wrap(ErrInsertRowCount,
				errs.WithContext("query", e.query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		// TODO:AFTER INSERT
	}
	return nil
}

// Close prepared statement.
func (e *insertExecutor) closePreparedStmt() {
	if e.prepared != nil {
		e.prepared.Close()
		e.gari().infoLog("insertExecutor.closePreparedStmt()")
	}
}

// Build INSERT SQL statement.
func (e *insertExecutor) buildQuery() string {
	tokens := []string{"INSERT", "INTO"}
	// Add table name.
	tokens = append(tokens, e.table.quotedName(), "(")
	// Add column names.
	for i, col := range e.columns {
		name := col.quotedName()
		if i < (len(e.columns) - 1) {
			name += ","
		}
		tokens = append(tokens, name)
	}
	// Add value placeholders.
	tokens = append(tokens, ")", "VALUES", "(")
	for i := range e.columns {
		ph := e.table.gari.placeholder(i)
		if i < (len(e.columns) - 1) {
			ph += ","
		}
		tokens = append(tokens, ph)
	}
	tokens = append(tokens, ");")
	query := strings.Join(tokens, " ")
	e.gari().debugLog("insertExecutor.buildQuery()",
		slog.String("query", query))
	return query
}

// Prepare INSERT SQL statement.
func (e *insertExecutor) prepareStmt() {
	var err error
	startedAt := time.Now()
	e.prepared, err = e.gari().db.Prepare(e.query)
	if err != nil {
		e.err = errs.Wrap(ErrInsertPrepare, errs.WithCause(err),
			errs.WithContext("query", e.query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari().errorLog(e.err.Error())
	}
	e.gari().infoLog("insertExecutor.prepareStmt()",
		slog.String("query", e.query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (e *insertExecutor) gari() *Gari {
	return e.table.gari
}
