package gari

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
)

type (
	// Executor to run raw SELECT SQL.
	rawSelectExecutor struct {
		gari  *Gari
		query string
		err   error // Error.
	}
)

var (
	ErrRawSelectExec        = errors.New("gari.ErrRawSelectExec")
	ErrRawSelectScan        = errors.New("gari.ErrRawSelectScan")
	ErrRawSelectRows        = errors.New("gari.ErrRawSelectRows")
	ErrRawSelectColumnTypes = errors.New("gari.ErrRawSelectColumnTypes")
)

// Create new selectExecutor.
func newRawSelectExecutor(g *Gari, query string) *rawSelectExecutor {
	query = strings.ReplaceAll(query, "\t", "")
	query = strings.ReplaceAll(query, "\n", "")
	e := rawSelectExecutor{
		gari:  g,
		query: query,
	}
	return &e
}

// Execute SELECT SQL query.
func (e *rawSelectExecutor) Exec(
	ctx context.Context, fn func(r *Record),
) error {
	if e.err != nil {
		return e.err
	}
	// Execute query.
	startedAt := time.Now()
	rows, err := e.gari.db.QueryContext(ctx, e.query)
	if err != nil {
		e.err = errs.Wrap(ErrRawSelectExec, errs.WithCause(err),
			errs.WithContext("query", e.query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari.errorLog(e.err.Error())
		return e.err
	}
	e.gari.infoLog("rawSelectExecutor.Exec",
		slog.String("query", e.query),
		slog.Duration("duration", time.Since(startedAt)))
	defer rows.Close()
	// Get column types from sql.Rows.
	types, err := rows.ColumnTypes()
	if err != nil {
		e.err = errs.Wrap(ErrRawSelectColumnTypes,
			errs.WithCause(err),
			errs.WithContext("query", e.query))
		e.gari.errorLog(e.err.Error())
		return e.err
	}
	// Build args.
	args := make([]any, len(types))
	values := make([]*value, len(types))
	for i, t := range types {
		v := newValueByColumnType(t)
		values[i] = v
		args[i] = v
	}
	// Scan rows.
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		// Scan values.
		err = rows.Scan(args...)
		if err != nil {
			err = errs.Wrap(ErrRawSelectScan, errs.WithCause(err),
				errs.WithContext("query", e.query))
			e.gari.errorLog(err.Error())
			return err
		}
		// Pass Record to func.
		r := newRecord(values)
		fn(r)
	}
	// Check error
	if rows.Err() != nil {
		err = errs.Wrap(ErrRawSelectRows, errs.WithCause(err),
			errs.WithContext("query", e.query))
		e.gari.errorLog(err.Error())
		return err
	}
	return nil
}
