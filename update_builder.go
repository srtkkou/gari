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
	// Builder to build UPDATE SQL.
	updateBuilder struct {
		table      *Table   // Pointer to table.
		fieldNames []string // Sorted field names of struct.
		query      string   // UPDATE SQL query.
	}
)

var (
	ErrUpdateBuilderBuildArgs = errors.New("gari.ErrUpdateBuilderBuildArgs")
)

// Create new updateBuilder instance.
func newUpdateBuilder(t *Table) *updateBuilder {
	b := updateBuilder{
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
	// Build UPDATE SQL query.
	b.query = b.buildQuery()
	fmt.Printf("UPDATE QUERY=%s\n", b.query)
	return &b
}

// Build SQL statement.
func (b *updateBuilder) buildQuery() string {
	var sb strings.Builder
	sb.WriteString("UPDATE ")
	// Build table statement.
	quote := b.table.gari.columnQuote
	sb.WriteString(quote)
	sb.WriteString(b.table.name)
	sb.WriteString(quote)
	sb.WriteString(" SET ")
	// Build values statement.
	for i, fieldName := range b.fieldNames {
		col := b.table.columns[fieldName]
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quote)
		sb.WriteString(col.name)
		sb.WriteString(quote)
		sb.WriteString(" = ?")
	}
	// Build WHERE statement.
	sb.WriteString(" WHERE ")
	sb.WriteString(quote)
	sb.WriteString("id")
	sb.WriteString(quote)
	sb.WriteString(" = ?;")
	return sb.String()
}

// Build argument of ptr.
func (b *updateBuilder) buildArgs(ptr any) ([]any, error) {
	// Convert to msgpack.
	blob, err := msgpack.Marshal(ptr)
	if err != nil {
		err = errs.Wrap(ErrUpdateBuilderBuildArgs,
			errs.WithCause(err))
		return nil, err
	}
	// Unmarshal msgpack.
	m, err := decodeToMap(blob)
	if err != nil {
		err = errs.Wrap(ErrUpdateBuilderBuildArgs,
			errs.WithCause(err),
			errs.WithContext("msgpack", blob))
		return nil, err
	}
	// Update timestamp.
	m["UpdatedAt"] = time.Now().UTC()
	// Build args.
	args := make([]any, len(b.fieldNames))
	for i, fieldName := range b.fieldNames {
		v := m[fieldName]
		fmt.Printf("args[%d] field=%s value=%v(%T)\n", i, fieldName, v, v)
		args[i] = v
	}
	return args, nil
}
