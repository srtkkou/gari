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
	ErrDeletePrepare      = errors.New("gari.ErrDeletePrepare")
	ErrDeleteValuePtr     = errors.New("gari.ErrDeleteValuePtr")
	ErrDeleteNoRecords    = errors.New("gari.ErrDeleteNoRecords")
	ErrDeleteExec         = errors.New("gari.ErrDeleteExec")
	ErrDeleteRowsAffected = errors.New("gari.ErrDeleteRowsAffected")
	ErrDeleteRowCount     = errors.New("gari.ErrDeleteRowCount")
)

// Create new deleteExecutor instance.
func newDeleteExecutor(t *Table) *deleteExecutor {
	e := deleteExecutor{
		table:   t,
		records: make([]*Record, 0),
	}
	return &e
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
		// Before delete callback.
		if e.table.BeforeDelete != nil {
			e.gari().debugLog("gari.Table.BeforeDelete")
			e.table.BeforeDelete(r)
		}
		// Execute prepared statement.
		query := e.query()
		args := r.args()
		startedAt := time.Now()
		result, err := e.prepared.ExecContext(ctx, args...)
		if err != nil {
			e.err = errs.Wrap(ErrDeleteExec, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("duration", time.Since(startedAt)))
			e.gari().errorLog(e.err.Error())
			return e.err
		}
		e.gari().infoLog("Execute DELETE SQL.",
			slog.String("query", query),
			slog.Any("args", args),
			slog.Duration("duration", time.Since(startedAt)))
		// Check row count.
		count, err := result.RowsAffected()
		if err != nil {
			e.err = errs.Wrap(ErrDeleteRowsAffected,
				errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
			e.gari().errorLog(e.err.Error())
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
		// After delete callback.
		if e.table.AfterDelete != nil {
			e.gari().debugLog("gari.Table.AfterDelete")
			e.table.AfterDelete(r)
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
	tokens := []string{"DELETE", "FROM"}
	// Add table name.
	tokens = append(tokens, e.table.quotedName())
	// Add WHERE statement.
	tokens = append(tokens, "WHERE")
	pkey := e.table.pkey.quotedName()
	tokens = append(tokens, pkey, "=")
	// Add id placeholder.
	bv := dialect.BindVar(0)
	tokens = append(tokens, bv)
	e.queryCache = strings.Join(tokens, " ") + dialect.QuerySuffix()
	return e.queryCache
}

// Prepare DELETE SQL statement.
func (e *deleteExecutor) prepareStmt() {
	startedAt := time.Now()
	var err error
	query := e.query()
	e.prepared, err = e.gari().db.Prepare(query)
	if err != nil {
		e.err = errs.Wrap(ErrDeletePrepare, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari().errorLog(e.err.Error())
	}
	e.gari().infoLog("deleteExecutor.prepareStmt()",
		slog.String("query", query),
		slog.Duration("duration", time.Since(startedAt)))
}

func (e *deleteExecutor) gari() *Gari {
	return e.table.gari
}
