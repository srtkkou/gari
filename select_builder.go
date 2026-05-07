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
	// Builder to build SELECT SQL.
	selectBuilder struct {
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

// Create new selectBuilder.
func newSelectBuilder(t *Table) *selectBuilder {
	b := selectBuilder{
		from:   t,
		values: make([]*value, len(t.columnNames)),
		orders: make([]order, 0),
		limit:  0,
	}
	// Initialize values from columns.
	for i, name := range t.columnNames {
		col := t.columns[name]
		b.values[i] = newValue(col)
	}
	return &b
}

// Add ORDER BY ASC statement.
func (b *selectBuilder) OrderAsc(name string) *selectBuilder {
	if b.err != nil {
		return b
	}
	order := order{columnName: name, isAsc: true}
	b.orders = append(b.orders, order)
	return b
}

// Add ORDER BY DESC statement.
func (b *selectBuilder) OrderDesc(name string) *selectBuilder {
	if b.err != nil {
		return b
	}
	order := order{columnName: name, isAsc: false}
	b.orders = append(b.orders, order)
	return b
}

// Execute SELECT SQL query.
func (b *selectBuilder) Exec(
	ctx context.Context, fn func(r *Record),
) error {
	if b.err != nil {
		return b.err
	}
	// Build SQL statement.
	query := b.buildSelect() + " " +
		b.buildOrderBy() + ";"
	// Execute query.
	db := b.from.gari.db
	startedAt := time.Now()
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		b.err = errs.Wrap(ErrSelectExec, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", time.Since(startedAt)))
		b.gari().errorLog(b.err.Error())
		return b.err
	}
	b.gari().infoLog("selectBuilder.Exec",
		slog.String("query", query),
		slog.Duration("duration", time.Since(startedAt)))
	defer rows.Close()
	// Scan rows.
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		// Scan values.
		args := make([]any, len(b.values))
		for i := range args {
			args[i] = b.values[i]
		}
		err = rows.Scan(args...)
		if err != nil {
			return err
		}
		// Pass Record to func.
		r := newRecord(b.values)
		fn(r)
	}
	// Check error
	if rows.Err() != nil {
		return err
	}
	return nil
}

// Build SELECT FROM SQL statement
func (b *selectBuilder) buildSelect() string {
	/*
		tokens := []string{"SELECT"}
		// Columns
		quotedTable := fmt.Sprintf("%s%s%s",
			quote, b.from.name, quote)
		for i, value := range b.values {
			quotedColumn :=
			alias :=
		}
		// Table
		tokens = append(tokens, "FROM", quotedTable)
		// Order

		// Limit
		if b.limit > 0 {
			tokens = append(tokens, "LIMIT")
			tokens = append(tokens, strconv.Itoa(b.limit))
		}
		return strings.Join(tokens, " ")
	*/
	var sb strings.Builder
	quote := b.from.gari.columnQuote
	sb.WriteString("SELECT ")
	for i, value := range b.values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quote)
		sb.WriteString(b.from.Name())
		sb.WriteString(quote)
		sb.WriteString(".")
		sb.WriteString(quote)
		sb.WriteString(value.column.name)
		sb.WriteString(quote)
	}
	sb.WriteString(" FROM ")
	sb.WriteString(quote)
	sb.WriteString(b.from.name)
	sb.WriteString(quote)
	return sb.String()
}

// Build ORDER BY SQL statement.
func (b *selectBuilder) buildOrderBy() string {
	var sb strings.Builder
	quote := b.from.gari.columnQuote
	sb.WriteString("ORDER BY ")
	for i, order := range b.orders {
		// Find column in table.
		col, ok := b.from.columns[order.columnName]
		if !ok {
			continue
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quote)
		sb.WriteString(b.from.name)
		sb.WriteString(quote)
		sb.WriteString(".")
		sb.WriteString(quote)
		sb.WriteString(col.name)
		sb.WriteString(quote)
		if order.isAsc {
			sb.WriteString(" ASC")
		} else {
			sb.WriteString(" DESC")
		}
	}
	return sb.String()
}

func (b *selectBuilder) gari() *Gari {
	return b.from.gari
}
