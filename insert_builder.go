package gari

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build INSERT SQL.
	insertBuilder struct {
		table      *Table // Pointer to table.
		fieldNames []string
		query      string
	}
)

var (
	ErrInsertBuilderBuildArgs = errors.New("gari.ErrInsertBuilderBuildArgs")
)

// Create new insertBuilder instance.
func newInsertBuilder(t *Table) *insertBuilder {
	b := insertBuilder{
		table:      t,
		fieldNames: make([]string, 0, len(t.columnNames)),
	}
	// Set field names except "id" column.
	for _, name := range t.columnNames {
		if name != "id" {
			col := t.columns[name]
			b.fieldNames = append(b.fieldNames, col.fieldName)
		}
	}
	// Sort field names.
	sort.Strings(b.fieldNames)
	// Build INSERT SQL query.
	b.query = b.buildQuery()
	fmt.Printf("INSERT FIELDS=[%v] QUERY=%s\n", b.fieldNames, b.query)
	return &b
}

// Build INSERT SQL statement.
func (b *insertBuilder) buildQuery() string {
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	// Build table token.
	quote := b.table.gari.columnQuote
	sb.WriteString(quote)
	sb.WriteString(b.table.name)
	sb.WriteString(quote)
	sb.WriteString(" (")
	// Build columns token.
	for i, fieldName := range b.fieldNames {
		col := b.table.columns[fieldName]
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quote)
		sb.WriteString(col.name)
		sb.WriteString(quote)
	}
	sb.WriteString(") VALUES (")
	// Build values token.
	for i := range b.fieldNames {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("?")
	}
	sb.WriteString(");")
	return sb.String()
}

// Build argument of ptr.
func (b *insertBuilder) buildArgs(ptr any) ([]any, error) {
	// Convert to msgpack.
	blob, err := msgpack.Marshal(ptr)
	if err != nil {
		err = errs.Wrap(ErrInsertBuilderBuildArgs,
			errs.WithCause(err))
		return []any{}, err
	}
	// Unmarshal msgpack.
	m, err := decodeToMap(blob)
	if err != nil {
		err = errs.Wrap(ErrInsertBuilderBuildArgs,
			errs.WithCause(err),
			errs.WithContext("msgpack", blob))
		return []any{}, err
	}
	// Update timestamps.
	ts := time.Now().UTC()
	m["CreatedAt"] = ts
	m["UpdatedAt"] = ts
	// Build args.
	args := make([]any, len(b.fieldNames))
	for i, fieldName := range b.fieldNames {
		v := m[fieldName]
		fmt.Printf("args[%d] field=%s value=%v(%T)\n", i, fieldName, v, v)
		args[i] = v
	}
	return args, nil
}
