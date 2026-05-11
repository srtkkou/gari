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
		table      *Table    // Pointer to table.
		columns    []*column // Slice of columns except primary key.
		queryCache string    // INSERT SQL query cache.
		prepared   *sql.Stmt // Prepared statement pointer.
		records    []*Record // Records to insert.
		err        error     // Error.
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
	e := &insertExecutor{
		table:   t,
		columns: make([]*column, 0, (len(t.columns) - 1)),
		records: make([]*Record, 0),
	}
	// Initialize columns.
	for _, col := range t.columns {
		// Skip primary key.
		if col.isPrimary {
			continue
		}
		e.columns = append(e.columns, col)
	}
	return e
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
		// Before INSERT hooks.
		if len(e.table.beforeInsertHooks) > 0 {
			e.gari().debugLog("gari.Table.BeforeInsert")
			for _, hook := range e.table.beforeInsertHooks {
				hook(r)
			}
		}
		// Execute prepared statement.
		query := e.query()
		args := r.args()
		startedAt := time.Now()
		result, err := e.prepared.ExecContext(ctx, args...)
		if err != nil {
			e.err = errs.Wrap(ErrInsertExec, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		e.gari().infoLog("Execute INSERT SQL.",
			slog.String("query", query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			e.err = errs.Wrap(ErrInsertRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		if count != 1 {
			e.err = errs.Wrap(ErrInsertRowCount,
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		// After INSERT hooks.
		if len(e.table.afterInsertHooks) > 0 {
			e.gari().debugLog("gari.Table.AfterInsert")
			for _, hook := range e.table.afterInsertHooks {
				hook(r)
			}
		}
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
func (e *insertExecutor) query() string {
	if len(e.queryCache) > 0 {
		return e.queryCache
	}
	dialect := e.gari().dialect
	var sb strings.Builder
	sb.WriteString(`INSERT INTO `)
	// Add table name.
	dialect.QuoteTable(&sb, e.table.name)
	sb.WriteString(` (`)
	// Add column names.
	for i, col := range e.columns {
		if i > 0 {
			sb.WriteString(`, `)
		}
		dialect.QuoteColumn(&sb, col.name)
	}
	// Add value placeholders.
	sb.WriteString(`) VALUES (`)
	for i := range e.columns {
		if i > 0 {
			sb.WriteString(`, `)
		}
		dialect.BindVar(&sb, i)
	}
	sb.WriteString(`)`)
	dialect.QuerySuffix(&sb)
	e.queryCache = sb.String()
	return e.queryCache
}

// Prepare INSERT SQL statement.
func (e *insertExecutor) prepareStmt() {
	var err error
	query := e.query()
	startedAt := time.Now()
	e.prepared, err = e.gari().db.Prepare(query)
	if err != nil {
		e.err = errs.Wrap(ErrInsertPrepare, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari().errorLog(e.err.Error())
	}
	e.gari().infoLog("insertExecutor.prepareStmt()",
		slog.String("query", query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (e *insertExecutor) gari() *Gari {
	return e.table.gari
}
