package gari

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	selectBuilder struct {
		ctx        context.Context
		db         *sql.DB
		from       *Table
		ptr        any
		cells      []*Cell
		orderStmts []string // ORDER BY statements.
		err        error
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
		cells:      make([]*Cell, len(t.columnNames)),
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
	stmt := b.orderByStmt(name) + " ASC"
	b.orderStmts = append(b.orderStmts, stmt)
	return b
}

// Add ORDER BY DESC statement.
func (b *selectBuilder) OrderDesc(name string) *selectBuilder {
	if b.err != nil {
		return b
	}
	stmt := b.orderByStmt(name) + " DESC"
	b.orderStmts = append(b.orderStmts, stmt)
	return b
}

// Select one item.
func (b *selectBuilder) First() error {
	if b.err != nil {
		return b.err
	}
	// Build SQL statement.
	query := b.selectStmt() +
		` ORDER BY ` + strings.Join(b.orderStmts, ", ") +
		` LIMIT 1;`
	fmt.Printf("----First----\n")
	fmt.Printf("--query=%s\n", query)
	// Execute query.
	row := b.db.QueryRowContext(b.ctx, query)
	err := row.Scan(b.scanArgs()...)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	} else if err != nil {
		return err
	}
	// Build msgpack bytes.
	blob, err := b.msgpackMapBytes()
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
	query := b.selectStmt() +
		` ORDER BY ` + strings.Join(b.orderStmts, ", ") + ";"
	fmt.Printf("----All----\n")
	fmt.Printf("--query=%s\n", query)
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
		fmt.Printf("--rows.Next():count=%d\n", count)
		// Scan row.
		err = rows.Scan(b.scanArgs()...)
		if err != nil {
			fmt.Printf("--err=%v\n", err)
			return err
		}
		// msgpack形式に変換
		mapBytes, err := b.msgpackMapBytes()
		if err != nil {
			return err
		}
		buf.Write(mapBytes)
	}
	// msgpackバイト列に配列ヘッダをつける。
	arrayBytes, err := msgpackArrayHeader(count)
	if err != nil {
		return err
	}
	// msgpackの組み立て
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

// Build SELECT FROM statement
func (b *selectBuilder) selectStmt() string {
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

// Build  ORDER BY statement.
func (b *selectBuilder) orderByStmt(name string) string {
	// Get column.
	col, ok := b.from.columns[name]
	if !ok {
		b.err = errs.Wrap(ErrColumnNotFound, errs.WithContext("name", name))
		return ""
	}
	// Build stmt.
	var sb strings.Builder
	quote := b.from.gari.columnQuote
	sb.WriteString(quote)
	sb.WriteString(b.from.name)
	sb.WriteString(quote)
	sb.WriteString(".")
	sb.WriteString(quote)
	sb.WriteString(col.name)
	sb.WriteString(quote)
	return sb.String()
}

// Convert type of sb.cells to []any.
func (b *selectBuilder) scanArgs() []any {
	args := make([]any, len(b.cells))
	for i, cell := range b.cells {
		args[i] = cell
	}
	return args
}

// msgpack map bytes.
func (b *selectBuilder) msgpackMapBytes() ([]byte, error) {
	var buf bytes.Buffer
	// Build map header.
	headerBytes, err := msgpackMapHeader(len(b.cells))
	if err != nil {
		return []byte{}, err
	}
	buf.Write(headerBytes)
	// Write key value pair.
	for _, cell := range b.cells {
		keyValueBytes, err := cell.MsgpackBytes()
		if err != nil {
			return []byte{}, err
		}
		buf.Write(keyValueBytes)
	}
	return buf.Bytes(), err
}
