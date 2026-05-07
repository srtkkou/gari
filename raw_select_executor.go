package gari

import (
	"context"
	"errors"
	"fmt"
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
	ErrRawSelectExec = errors.New("gari.ErrRawSelectExec")
	ErrRawSelectScan = errors.New("gari.ErrRawSelectScan")
	ErrRawSelectRows = errors.New("gari.ErrRawSelectRows")
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
	ctx context.Context,
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
	e.gari.infoLog("rawSelectExecutor.Exec",
		slog.String("query", e.query),
		slog.Duration("duration", time.Since(startedAt)))
	defer rows.Close()
	// Build args.
	fmt.Println("rawSelectExecutor.Exec() Get types")
	types, err := rows.ColumnTypes()
	if err != nil {
		e.err = errs.Wrap(ErrSelectExec, errs.WithCause(err),
			errs.WithContext("query", e.query))
		e.gari.errorLog(e.err.Error())
		return e.err
	}
	fmt.Println("rawSelectExecutor.Exec() Get types DONE")
	args := make([]any, len(types))
	values := make([]*value, len(types))
	for i := range values {
		length, _ := types[i].Length()
		precision, scale, _ := types[i].DecimalSize()
		nullable, _ := types[i].Nullable()
		v := &value{
			name:      types[i].Name(),
			typeName:  types[i].DatabaseTypeName(),
			length:    length,
			precision: precision,
			scale:     scale,
			nullable:  nullable,
		}
		e.gari.debugLog(fmt.Sprintf("value name=%s type=%s length=%d precision=%d scale=%d nullable=%v", v.name, v.typeName, v.length, v.precision, v.scale, v.nullable))
		values[i] = v
		args[i] = v
	}
	fmt.Println("rawSelectExecutor.Exec() init args DONE")
	// Scan rows.
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		// Scan values.
		fmt.Println("rawSelectExecutor.Exec() rows.Scan")
		err = rows.Scan(args...)
		if err != nil {
			err = errs.Wrap(ErrRawSelectScan, errs.WithCause(err),
				errs.WithContext("query", e.query))
			e.gari.errorLog(err.Error())
			return err
		}
		fmt.Println("rawSelectExecutor.Exec() rows.Scan DONE")
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
