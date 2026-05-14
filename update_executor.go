package gari

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/goark/errs"
)

// Executor to run UPDATE SQL.
type updateExecutor struct {
	table      *Table    // Pointer to table.
	columns    []*column // Slice of column pointers.
	queryCache string    // UPDATE SQL query cache.
	prepared   *sql.Stmt // Prepared statement pointer.
	records    []*Record // Records to update.
	err        error     // Error.
}

var (
	ErrUpdateValuePtr     = errors.New("gari.ErrUpdateValuePtr")
	ErrUpdateNoRecords    = errors.New("gari.ErrUpdateNoRecords")
	ErrUpdateRowsAffected = errors.New("gari.ErrUpdateRowsAffected")
	ErrUpdateRowCount     = errors.New("gari.ErrUpdateRowCount")
)

// Create new updateExecutor instance.
func newUpdateExecutor(t *Table) *updateExecutor {
	e := &updateExecutor{
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
	return e
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
			fieldName := e.table.mapper.FieldNameOf(col.Name)
			values[i].raw = m[fieldName]
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
		// Before UPDATE hooks.
		if len(e.table.beforeUpdateHooks) > 0 {
			e.gari().debugLog("gari.Table.BeforeUpdate")
			for _, hook := range e.table.beforeUpdateHooks {
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
			e.err = errs.Wrap(ErrUpdateRowCount,
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		// After UPDATE hooks.
		if len(e.table.afterUpdateHooks) > 0 {
			e.gari().debugLog("gari.Table.AfterUpdate")
			for _, hook := range e.table.afterUpdateHooks {
				hook(r)
			}
		}
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
func (e *updateExecutor) query() string {
	if len(e.queryCache) > 0 {
		return e.queryCache
	}
	dialect := e.gari().dialect
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	// Add table name.
	dialect.QuoteTable(&sb, e.table.Name)
	sb.WriteString(" SET ")
	// Add column names and placeholders.
	count := 0
	for _, col := range e.columns {
		if count > 0 {
			sb.WriteString(`, `)
		}
		dialect.QuoteColumn(&sb, col.Name)
		sb.WriteString(` = `)
		dialect.BindVar(&sb, count)
		count++
	}
	// Add WHERE statement.
	sb.WriteString(` WHERE `)
	dialect.QuoteColumn(&sb, e.table.pkey.Name)
	sb.WriteString(` = `)
	dialect.BindVar(&sb, count)
	dialect.QuerySuffix(&sb)
	e.queryCache = sb.String()
	return e.queryCache
}

// Prepare UPDATE SQL statement.
func (e *updateExecutor) prepareStmt() {
	e.prepared, e.err = e.gari().prepare(e.query())
}

func (e *updateExecutor) gari() *Gari {
	return e.table.gari
}
