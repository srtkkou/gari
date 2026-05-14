package gari

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/goark/errs"
)

type (
	// Executor to run DELETE SQL.
	deleteExecutor struct {
		table      *Table    // Pointer to table.
		queryCache string    // DELETE SQL query cache.
		prepared   *sql.Stmt // Prepared statement pointer.
		records    []*Record // Records to update.
		err        error     // Error.
	}
)

var (
	ErrDeleteValuePtr     = errors.New("gari.ErrDeleteValuePtr")
	ErrDeleteNoRecords    = errors.New("gari.ErrDeleteNoRecords")
	ErrDeleteRowsAffected = errors.New("gari.ErrDeleteRowsAffected")
	ErrDeleteRowCount     = errors.New("gari.ErrDeleteRowCount")
)

// Create new deleteExecutor instance.
func newDeleteExecutor(t *Table) *deleteExecutor {
	e := &deleteExecutor{
		table:   t,
		records: make([]*Record, 0),
	}
	return e
}

// Add values to dalete.
func (e *deleteExecutor) Values(ptrs ...any) *deleteExecutor {
	if e.err != nil {
		return e
	}
	for _, ptr := range ptrs {
		// Build map of values from pointer to struct.
		m, err := structPtrToMap(ptr)
		if err != nil {
			e.err = errs.Wrap(ErrDeleteValuePtr,
				errs.WithCause(err),
				errs.WithContext("ptr", ptr))
			e.gari().errorLog(e.err.Error())
			return e
		}
		// Build record.
		values := make([]*value, 1)
		values[0] = newValueByColumn(e.table.pkey)
		values[0].raw = m[e.table.pkey.fieldName]
		r := newRecord(values)
		e.records = append(e.records, r)
	}
	return e
}

// Execute UPDATE SQL statement.
func (e *deleteExecutor) Exec(ctx context.Context) error {
	defer func() {
		e.records = make([]*Record, 0)
	}()
	// Check if error exists.
	if e.err != nil {
		return e.err
	}
	// Check if records are not empty.
	if len(e.records) == 0 {
		e.err = errs.Wrap(ErrDeleteNoRecords,
			errs.WithContext("query", e.query),
			errs.WithContext("records", 0))
		e.gari().errorLog(e.err.Error())
		return e.err
	}
	for _, r := range e.records {
		// Before DELETE hooks.
		if len(e.table.beforeDeleteHooks) > 0 {
			e.gari().debugLog("gari.Table.BeforeDelete")
			for _, hook := range e.table.beforeDeleteHooks {
				hook(r)
			}
		}
		// Execute prepared statement.
		query := e.query()
		args := r.args()
		var count int64
		_, count, e.err = e.gari().execPrepared(ctx, e.prepared, args...)
		if e.err != nil {
			return e.err
		}
		if count != 1 {
			e.err = errs.Wrap(ErrDeleteRowCount,
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		// After DELETE hooks.
		if len(e.table.afterDeleteHooks) > 0 {
			e.gari().debugLog("gari.Table.AfterDelete")
			for _, hook := range e.table.afterDeleteHooks {
				hook(r)
			}
		}
	}
	return nil
}

// Close prepared statement.
func (e *deleteExecutor) closePreparedStmt() {
	if e.prepared != nil {
		e.prepared.Close()
	}
}

// Build SQL statement.
func (e *deleteExecutor) query() string {
	if len(e.queryCache) > 0 {
		return e.queryCache
	}
	dialect := e.gari().dialect
	var sb strings.Builder
	sb.WriteString("DELETE FROM ")
	// Add table name.
	dialect.QuoteTable(&sb, e.table.Name)
	// Add WHERE statement.
	sb.WriteString(" WHERE ")
	dialect.QuoteColumn(&sb, e.table.pkey.name)
	sb.WriteString(" = ")
	dialect.BindVar(&sb, 0)
	dialect.QuerySuffix(&sb)
	e.queryCache = sb.String()
	return e.queryCache
}

// Prepare DELETE SQL statement.
func (e *deleteExecutor) prepareStmt() {
	e.prepared, e.err = e.gari().prepare(e.query())
}

func (e *deleteExecutor) gari() *Gari {
	return e.table.gari
}
