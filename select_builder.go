package gari

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build SELECT SQL.
	selectBuilder struct {
		ctx        context.Context // Context.
		db         *sql.DB         // Pointer to database pool.
		from       *Table          // Main table to select from.
		ptr        any             // Destination to store selected models.
		cells      []*cell
		orders     []order  // ORDER BY statements.
		orderStmts []string // ORDER BY statements.
		err        error    // Error.
	}
	// ORDER BY statements.
	order struct {
		columnName string // Column name.
		isAsc      bool   // ASC or DESC.
	}
	// Interface to treat sql.Row and sql.Rows.
	row interface {
		Scan(dest ...any) error
	}
)

var (
	// Column not found from name.
	ErrColumnNotFound = errors.New("ErrColumnNotFound")
)

// Create new selectBuilder.
func newSelectBuilder(
	ctx context.Context, db *sql.DB, t *Table, ptr any,
) *selectBuilder {
	// Initialize selectBuilder.
	b := selectBuilder{
		ctx:        ctx,
		db:         db,
		from:       t,
		ptr:        ptr,
		cells:      make([]*cell, len(t.columnNames)),
		orders:     make([]order, 0),
		orderStmts: make([]string, 0),
	}
	// Initialize cells from columns.
	for i, name := range t.columnNames {
		col := t.columns[name]
		b.cells[i] = newCell(col)
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

// Select one item.
func (b *selectBuilder) First() error {
	if b.err != nil {
		return b.err
	}
	// Build SQL statement.
	query := b.buildSelect() + " " +
		b.buildOrderBy() + ` LIMIT 1;`
	fmt.Printf("--SelectFirst query=%s\n", query)
	// Execute query.
	row := b.db.QueryRowContext(b.ctx, query)
	// Convert result to msgpack.
	blob, err := b.rowToMsgpack(row)
	if err != nil {
		return err
	}
	fmt.Printf("msgpack=%#x\n", blob)
	// Unmarshal model.
	err = msgpack.Unmarshal(blob, b.ptr)
	if err != nil {
		return err
	}
	return nil
}

func (b *selectBuilder) All() error {
	if b.err != nil {
		return b.err
	}
	// Build SQL statement.
	query := b.buildSelect() + " " +
		b.buildOrderBy() + ";"
	fmt.Printf("--SelectAll query=%s\n", query)
	// Execute query.
	rows, err := b.db.QueryContext(b.ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Scan rows.
	var buf bytes.Buffer
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		// Convert result to msgpack.
		blob, err := b.rowToMsgpack(rows)
		if err != nil {
			return err
		}
		buf.Write(blob)
	}
	// Add msgpack array header.
	arrayBytes, err := msgpackArrayHeader(count)
	if err != nil {
		return err
	}
	// Build msgpack.
	blob := append(arrayBytes, buf.Bytes()...)
	fmt.Printf("msgpack=%#x\n", blob)
	// Unmarshal model
	err = msgpack.Unmarshal(blob, b.ptr)
	if err != nil {
		fmt.Printf("--UnmarshalErr=%v\n", err)
		return err
	}
	// Check error
	if rows.Err() != nil {
		return err
	}
	return nil
}

// Build SELECT FROM SQL statement
func (b *selectBuilder) buildSelect() string {
	var sb strings.Builder
	quote := b.from.gari.columnQuote
	sb.WriteString("SELECT ")
	for i, cell := range b.cells {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quote)
		sb.WriteString(b.from.Name())
		sb.WriteString(quote)
		sb.WriteString(".")
		sb.WriteString(quote)
		sb.WriteString(cell.column.name)
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

func (b *selectBuilder) rowToMsgpack(row row) ([]byte, error) {
	var buf bytes.Buffer
	// Add msgpack map header.
	size := len(b.cells)
	headerBytes, err := msgpackMapHeader(size)
	if err != nil {
		return nil, err
	}
	buf.Write(headerBytes)
	// Scan values.
	args := make([]any, size)
	for i := range args {
		args[i] = b.cells[i]
	}
	err = row.Scan(args...)
	if err != nil {
		return nil, err
	}
	// Add msgpack encoded keys and values.
	for _, cell := range b.cells {
		blob, err := cell.Encode()
		if err != nil {
			return nil, err
		}
		buf.Write(blob)
	}
	return buf.Bytes(), nil
}
