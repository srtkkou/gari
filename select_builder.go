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
func newSelectBuilder(t *Table, ptr any) *selectBuilder {
	// Initialize selectBuilder.
	sb := selectBuilder{
		from:       t,
		ptr:        ptr,
		cells:      make([]*Cell, len(t.columnNames)),
		orderStmts: make([]string, 0),
	}
	// Initialize cells from columns.
	for i, name := range t.columnNames {
		col := t.columns[name]
		sb.cells[i] = newCell(col)
	}
	return &sb
}

// Add ORDER BY ASC statement.
func (sb *selectBuilder) OrderAsc(name string) *selectBuilder {
	if sb.err != nil {
		return sb
	}
	stmt := sb.orderByStmt(name) + " ASC"
	sb.orderStmts = append(sb.orderStmts, stmt)
	return sb
}

// Add ORDER BY DESC statement.
func (sb *selectBuilder) OrderDesc(name string) *selectBuilder {
	if sb.err != nil {
		return sb
	}
	stmt := sb.orderByStmt(name) + " DESC"
	sb.orderStmts = append(sb.orderStmts, stmt)
	return sb
}

// Select one item.
func (sb *selectBuilder) First(
	ctx context.Context, db *sql.DB,
) error {
	if sb.err != nil {
		return sb.err
	}
	// Build SQL statement.
	query := sb.selectStmt() +
		` ORDER BY ` + strings.Join(sb.orderStmts, ", ") +
		` LIMIT 1;`
	fmt.Printf("----First----\n")
	fmt.Printf("--query=%s\n", query)
	// Execute query.
	row := db.QueryRowContext(ctx, query)
	err := row.Scan(sb.scanArgs()...)
	if errors.Is(err, sql.ErrNoRows) {
		return err
	} else if err != nil {
		return err
	}
	// Build msgpack bytes.
	blob, err := sb.msgpackMapBytes()
	if err != nil {
		return err
	}
	fmt.Printf("msgpack=%#x\n", blob)
	// Unmarshal model.
	err = msgpack.Unmarshal(blob, sb.ptr)
	if err != nil {
		return err
	}
	return nil
}

func (sb *selectBuilder) All(
	ctx context.Context, db *sql.DB,
) error {
	if sb.err != nil {
		return sb.err
	}
	// Build SQL statement.
	query := sb.selectStmt() +
		` ORDER BY ` + strings.Join(sb.orderStmts, ", ") + ";"
	fmt.Printf("----All----\n")
	fmt.Printf("--query=%s\n", query)
	// Execute query.
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return err
	}
	defer rows.Close()
	// Scan rows.
	var b bytes.Buffer
	count := 0
	for rows.Next() {
		// Count up.
		count += 1
		fmt.Printf("--rows.Next():count=%d\n", count)
		// Scan row.
		err = rows.Scan(sb.scanArgs()...)
		if err != nil {
			fmt.Printf("--err=%v\n", err)
			return err
		}
		// msgpack形式に変換
		mapBytes, err := sb.msgpackMapBytes()
		if err != nil {
			return err
		}
		b.Write(mapBytes)
	}
	// msgpackバイト列に配列ヘッダをつける。
	arrayBytes, err := msgpackArrayHeader(count)
	if err != nil {
		return err
	}
	// msgpackの組み立て
	blob := append(arrayBytes, b.Bytes()...)
	fmt.Printf("msgpack=%#x\n", blob)
	// Unmarshal model
	err = msgpack.Unmarshal(blob, sb.ptr)
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
func (sb *selectBuilder) selectStmt() string {
	var b strings.Builder
	colQuote := sb.from.gari.columnQuote
	b.WriteString("SELECT ")
	for i, cell := range sb.cells {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(colQuote)
		b.WriteString(sb.from.Name())
		b.WriteString(colQuote)
		b.WriteString(".")
		b.WriteString(colQuote)
		b.WriteString(cell.column.name)
		b.WriteString(colQuote)
	}
	b.WriteString(" FROM ")
	b.WriteString(colQuote)
	b.WriteString(sb.from.Name())
	b.WriteString(colQuote)
	return b.String()
}

// Build  ORDER BY statement.
func (sb *selectBuilder) orderByStmt(name string) string {
	// Get column.
	col, ok := sb.from.columns[name]
	if !ok {
		sb.err = errs.Wrap(ErrColumnNotFound, errs.WithContext("name", name))
		return ""
	}
	// Build stmt.
	var b strings.Builder
	colQuote := sb.from.gari.columnQuote
	b.WriteString(colQuote)
	b.WriteString(sb.from.Name())
	b.WriteString(colQuote)
	b.WriteString(".")
	b.WriteString(colQuote)
	b.WriteString(col.name)
	b.WriteString(colQuote)
	return b.String()
}

// Convert type of sb.cells to []any.
func (sb *selectBuilder) scanArgs() []any {
	args := make([]any, len(sb.cells))
	for i, cell := range sb.cells {
		args[i] = cell
	}
	return args
}

// msgpack map bytes.
func (sb *selectBuilder) msgpackMapBytes() ([]byte, error) {
	var b bytes.Buffer
	// Build map header.
	headerBytes, err := msgpackMapHeader(len(sb.cells))
	if err != nil {
		return []byte{}, err
	}
	b.Write(headerBytes)
	// Write key value pair.
	for _, cell := range sb.cells {
		keyValueBytes, err := cell.MsgpackBytes()
		if err != nil {
			return []byte{}, err
		}
		b.Write(keyValueBytes)
	}
	return b.Bytes(), err
}
