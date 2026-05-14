package gari

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/goark/errs"
)

// Executor to run raw SELECT SQL.
type selectExecutor struct {
	gari  *Gari
	ptr   any
	query string
	err   error // Error.
}

var (
	ErrSelectExec        = errors.New("gari.ErrSelectExec")
	ErrSelectScan        = errors.New("gari.ErrSelectScan")
	ErrSelectRows        = errors.New("gari.ErrSelectRows")
	ErrSelectColumnTypes = errors.New("gari.ErrSelectColumnTypes")
)

// Create new selectExecutor.
func newSelectExecutor(
	g *Gari, ptr any, query string,
) *selectExecutor {
	e := selectExecutor{
		gari:  g,
		ptr:   ptr,
		query: squashSpace(query),
	}
	return &e
}

func (e *selectExecutor) AssignTo(ptrs any) *selectExecutor {
	return e
}

// Execute SELECT SQL query.
func (e *selectExecutor) Exec(
	ctx context.Context, fn func(r *Record),
) error {
	if e.err != nil {
		return e.err
	}
	// Execute query.
	startedAt := time.Now()
	rows, err := e.gari.db.QueryContext(ctx, e.query)
	if err != nil {
		e.err = errs.Wrap(ErrSelectExec, errs.WithCause(err),
			errs.WithContext("query", e.query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari.errorLog(e.err.Error())
		return e.err
	}
	e.gari.infoLog("selectExecutor.Exec",
		slog.String("query", e.query),
		slog.Duration("duration", time.Since(startedAt)))
	defer rows.Close()
	// Get column types from sql.Rows.
	types, err := rows.ColumnTypes()
	if err != nil {
		e.err = errs.Wrap(ErrSelectColumnTypes,
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
			err = errs.Wrap(ErrSelectScan, errs.WithCause(err),
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
		err = errs.Wrap(ErrSelectRows, errs.WithCause(err),
			errs.WithContext("query", e.query))
		e.gari.errorLog(err.Error())
		return err
	}
	return nil
}
