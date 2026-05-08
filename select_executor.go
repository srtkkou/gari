package gari

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/goark/errs"
)

type (
	// Executor to run SELECT SQL.
	selectExecutor struct {
		from    *Table // Main table to select from.
		columns []*column
		values  []*value
		err     error // Error.

		//		joins  []join  // JOIN statements.
		orders []order // ORDER BY statements.
		offset int     // OFFSET statement.
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
		column *column // Column.
		isAsc  bool    // ASC or DESC.
	}
)

var (
	ErrOrderColumnNotFound = errors.New("ErrOrderColumnNotFound")
	ErrSelectExec          = errors.New("gari.ErrSelectExec")
	ErrSelectColumnTypes   = errors.New("gari.ErrSelectColumnTypes")
	ErrSelectScan          = errors.New("gari.ErrSelectScan")
)

// Create new selectExecutor.
func newSelectExecutor(t *Table) *selectExecutor {
	e := selectExecutor{
		from:    t,
		columns: make([]*column, 0, len(t.columns)),
		//values:  make([]*value, len(t.columns)),
		orders: make([]order, 0),
		offset: -1,
		limit:  -1,
	}
	// Copy columns from table.
	e.columns = append(e.columns, t.columns...)
	/*
		// Initialize values from columns.
		for i, col := range t.columns {
			e.values[i] = newValueByColumn(col)
		}
	*/
	return &e
}

// Add ORDER BY ASC statement.
func (e *selectExecutor) OrderAsc(name string) *selectExecutor {
	return e.addOrder(name, true)
}

// Add ORDER BY DESC statement.
func (e *selectExecutor) OrderDesc(name string) *selectExecutor {
	return e.addOrder(name, false)
}

// Execute SELECT SQL query.
func (e *selectExecutor) Exec(
	ctx context.Context, fn func(r *Record),
) error {
	if e.err != nil {
		return e.err
	}
	// Build SQL statement.
	query := e.buildQuery()
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
	// Get column types from sql.Rows.
	types, err := rows.ColumnTypes()
	if err != nil {
		e.err = errs.Wrap(ErrSelectColumnTypes,
			errs.WithCause(err),
			errs.WithContext("query", query))
		e.gari().errorLog(e.err.Error())
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
				errs.WithContext("query", query))
			e.gari().errorLog(err.Error())
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
func (e *selectExecutor) buildQuery() string {
	tokens := []string{"SELECT"}
	// Column names to select..
	for i, column := range e.columns {
		// Column name.
		name := column.quotedFullName()
		tokens = append(tokens, name)
		// Alias.
		alias := column.fullName()
		alias = e.gari().quoteIdentifier(alias)
		if i < (len(e.columns) - 1) {
			alias += ","
		}
		tokens = append(tokens, "AS", alias)
	}
	// FROM table statement.
	tokens = append(tokens, "FROM")
	tokens = append(tokens, e.from.quotedName())
	// WHERE statement.
	// ORDER BY statement.
	if len(e.orders) > 0 {
		tokens = append(tokens, "ORDER", "BY")
		for i, order := range e.orders {
			// Column name.
			name := order.column.quotedFullName()
			tokens = append(tokens, name)
			// ASC or DESC.
			var ascDesc string
			if order.isAsc {
				ascDesc = "ASC"
			} else {
				ascDesc = "DESC"
			}
			if i < (len(e.orders) - 1) {
				ascDesc += ","
			}
			tokens = append(tokens, ascDesc)
		}
	}
	// OFFSET statement.
	if e.offset > 0 {
		tokens = append(tokens, "OFFSET")
		tokens = append(tokens, strconv.Itoa(e.offset))
	}
	// LIMIT statement.
	if e.limit > 0 {
		tokens = append(tokens, "LIMIT")
		tokens = append(tokens, strconv.Itoa(e.limit))
	}
	// Build query.
	query := strings.Join(tokens, " ") + ";"
	e.gari().debugLog("selectExecutor.buildQuery()",
		slog.String("query", query))
	return query
}

// Add ORDER BY statement.
func (e *selectExecutor) addOrder(name string, isAsc bool) *selectExecutor {
	if e.err != nil {
		return e
	}
	// Find column in tables.
	col := e.from.column(name)
	// TODO: Find column in joined tables.
	if col == nil {
		e.err = errs.Wrap(ErrOrderColumnNotFound,
			errs.WithContext("name", name))
		e.gari().errorLog(e.err.Error())
		return e
	}
	order := order{column: col, isAsc: isAsc}
	e.orders = append(e.orders, order)
	return e
}

func (e *selectExecutor) gari() *Gari {
	return e.from.gari
}
