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
	// Executor to run SELECT SQL.
	selectExecutor struct {
		from   *Table // Main table to select from.
		values []*value
		err    error // Error.

		//		joins  []join  // JOIN statements.
		orders []order // ORDER BY statements.
		limit  int     // LIMIT statement.
	}
	// JOIN statements.
	//	join struct {
	//		table    *Table // Pointer to table.
	//		modifier string // JOIN modifier.
	//		on       string // ON condition.
	//	}
	// ORDER BY statements.
	order struct {
		columnName string // Column name.
		isAsc      bool   // ASC or DESC.
	}
)

var (
	ErrSelectExec     = errors.New("gari.ErrSelectExec")
	ErrColumnNotFound = errors.New("ErrColumnNotFound")
)

// Create new selectExecutor.
func newSelectExecutor(t *Table) *selectExecutor {
	e := selectExecutor{
		from:   t,
		values: make([]*value, len(t.columns)),
		orders: make([]order, 0),
		limit:  0,
	}
	// Initialize values from columns.
	for i, col := range t.columns {
		e.values[i] = newValue(col)
	}
	return &e
}

// Add ORDER BY ASC statement.
func (e *selectExecutor) OrderAsc(name string) *selectExecutor {
	if e.err != nil {
		return e
	}
	order := order{columnName: name, isAsc: true}
	e.orders = append(e.orders, order)
	return e
}

// Add ORDER BY DESC statement.
func (e *selectExecutor) OrderDesc(name string) *selectExecutor {
	if e.err != nil {
		return e
	}
	order := order{columnName: name, isAsc: false}
	e.orders = append(e.orders, order)
	return e
}

// Execute SELECT SQL query.
func (e *selectExecutor) Exec(
	ctx context.Context, fn func(r *Record),
) error {
	if e.err != nil {
		return e.err
	}
	// Build SQL statement.
	query := e.buildSelect() + " " +
		e.buildOrderBy() + ";"
	// Execute query.
	startedAt := time.Now()
	rows, err := e.gari().db.QueryContext(ctx, query)
	if err != nil {
		e.err = errs.Wrap(ErrSelectExec, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", time.Since(startedAt)))
		e.gari().errorLog(e.err.Error())
		return e.err
	}
	e.gari().infoLog("selectExecutor.Exec",
		slog.String("query", query),
		slog.Duration("duration", time.Since(startedAt)))
	defer rows.Close()
	// Scan rows.
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		// Scan values.
		args := make([]any, len(e.values))
		for i := range args {
			args[i] = e.values[i]
		}
		err = rows.Scan(args...)
		if err != nil {
			return err
		}
		// Pass Record to func.
		r := newRecord(e.values)
		fn(r)
	}
	// Check error
	if rows.Err() != nil {
		return err
	}
	return nil
}

// Build SELECT FROM SQL statement
func (e *selectExecutor) buildSelect() string {
	/*
		tokens := []string{"SELECT"}
		// Columns
		quotedTable := fmt.Sprintf("%s%s%s",
			quote, e.from.name, quote)
		for i, value := range e.values {
			quotedColumn :=
			alias :=
		}
		// Table
		tokens = append(tokens, "FROM", quotedTable)
		// Order

		// Limit
		if e.limit > 0 {
			tokens = append(tokens, "LIMIT")
			tokens = append(tokens, strconv.Itoa(e.limit))
		}
		return strings.Join(tokens, " ")
	*/
	var sb strings.Builder
	table := e.gari().quoteIdentifier(e.from.name)
	sb.WriteString("SELECT ")
	for i, value := range e.values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(table)
		sb.WriteString(".")
		colName := e.gari().quoteIdentifier(value.column.name)
		sb.WriteString(colName)
	}
	sb.WriteString(" FROM ")
	sb.WriteString(table)
	return sb.String()
}

// Build ORDER BY SQL statement.
func (e *selectExecutor) buildOrderBy() string {
	var sb strings.Builder
	table := e.gari().quoteIdentifier(e.from.name)
	sb.WriteString("ORDER BY ")
	for i, order := range e.orders {
		// Find column in table.
		col := e.from.column(order.columnName)
		if col == nil {
			continue
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(table)
		sb.WriteString(".")
		colName := e.gari().quoteIdentifier(col.name)
		sb.WriteString(colName)
		if order.isAsc {
			sb.WriteString(" ASC")
		} else {
			sb.WriteString(" DESC")
		}
	}
	return sb.String()
}

func (e *selectExecutor) gari() *Gari {
	return e.from.gari
}
