package gari

import (
	"context"
	"errors"
	"fmt"
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
	ErrSelectStructNew   = errors.New("gari.ErrSelectStructNew")
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
	// TODO: Create slice.
	if e.ptr == nil {
		makeSlice(e.ptr)
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
		// Build struct to store values.
		itemPtr, err := newStruct(e.ptr)
		if err != nil {
			err = errs.Wrap(ErrSelectStructNew,
				errs.WithCause(err),
				errs.WithContext("query", e.query))
			e.gari.errorLog(err.Error())
			return err
		}
		fmt.Printf("NewStruct=%v(%T)\n", itemPtr, itemPtr)
		// Copy values into struct.
		r := newRecord(values)
		//lint:ignore SA4009 Overwrite ptr is needed.
		eachField(itemPtr, func(name string, ptr any) {
			colName := e.gari.mapper.ColumnNameOf(name)
			value, ok := r.valueByName(colName)
			if ok {
				ptr = &value.raw
				fmt.Printf("name=%s colName=%s ptr=%v(%T) Value=%s\n", name, colName, ptr, ptr, value)
			}
		})
		// Append to slice.
		appendToSlice(e.ptr, itemPtr)
		//		fn(r)
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
