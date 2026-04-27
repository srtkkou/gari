package gari

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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
	tokens := []string{
		"UPDATE", b.table.name, "SET",
	}
	// Build values token.
	values := make([]string, len(b.fieldNames))
	for i, fieldName := range b.fieldNames {
		col := b.table.columns[fieldName]
		values[i] = fmt.Sprintf("%s = ?", col.name)
	}
	tokens = append(tokens, strings.Join(values, ", "))
	tokens = append(tokens, "WHERE ID = ?;")
	return strings.Join(tokens, " ")
}

// Build argument of ptr.
func (b *updateBuilder) buildArgs(ptr any) ([]any, error) {
	// Convert to msgpack.
	blob, err := msgpack.Marshal(ptr)
	if err != nil {
		err = errs.Wrap(ErrUpdateBuilderBuildArgs,
			errs.WithCause(err))
		return []any{}, err
	}
	// Unmarshal msgpack.
	var m map[string]any
	err = msgpack.Unmarshal(blob, &m)
	if err != nil {
		err = errs.Wrap(ErrUpdateBuilderBuildArgs,
			errs.WithCause(err),
			errs.WithContext("msgpack", blob))
		return []any{}, err
	}
	// Build args.
	quote := b.table.gari.stringQuote
	args := make([]any, len(b.fieldNames))
	for i, fieldName := range b.fieldNames {
		v := m[fieldName]
		switch tv := v.(type) {
		case nil:
			args[i] = "NULL"
		case string:
			args[i] = quote + tv + quote
		case time.Time:
			args[i] = quote + tv.Format("20060102 15:06:07.999999") + quote
		case int8:
			args[i] = strconv.FormatInt(int64(tv), 10)
		case int16:
			args[i] = strconv.FormatInt(int64(tv), 10)
		case int32:
			args[i] = strconv.FormatInt(int64(tv), 10)
		case int64:
			args[i] = strconv.FormatInt(tv, 10)
		case uint16:
			args[i] = strconv.FormatUint(uint64(tv), 10)
		case uint32:
			args[i] = strconv.FormatUint(uint64(tv), 10)
		case uint64:
			args[i] = strconv.FormatUint(tv, 10)
		default:
			args[i] = "ERR"
		}
	}
	return args, nil
}
