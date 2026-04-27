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
		fieldNames: make([]string, len(t.columnNames), 0),
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
	tokens := []string{
		"INSERT", "INTO", b.table.name, "VALUES",
	}
	// Build values token.
	values := make([]string, len(b.fieldNames))
	for i := range b.fieldNames {
		values[i] = "?"
	}
	valuesToken := fmt.Sprintf("(%s);",
		strings.Join(values, ", "))
	tokens = append(tokens, valuesToken)
	return strings.Join(tokens, " ")
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
	var m map[string]any
	err = msgpack.Unmarshal(blob, &m)
	if err != nil {
		err = errs.Wrap(ErrInsertBuilderBuildArgs,
			errs.WithCause(err),
			errs.WithContext("msgpack", blob))
		return []any{}, err
	}
	// Build args.
	quote := b.table.gari.stringQuote
	args := make([]any, len(b.fieldNames), 0)
	for _, fieldName := range b.fieldNames {
		v := m[fieldName]
		switch tv := v.(type) {
		case nil:
			args = append(args, "NULL")
		case string:
			args = append(args, quote+tv+quote)
		case time.Time:
			args = append(args, quote+tv.Format("20060102 15:06:07.999999")+quote)
		case int8:
			args = append(args, strconv.FormatInt(int64(tv), 10))
		case int16:
			args = append(args, strconv.FormatInt(int64(tv), 10))
		case int32:
			args = append(args, strconv.FormatInt(int64(tv), 10))
		case int64:
			args = append(args, strconv.FormatInt(tv, 10))
		case uint16:
			args = append(args, strconv.FormatUint(uint64(tv), 10))
		case uint32:
			args = append(args, strconv.FormatUint(uint64(tv), 10))
		case uint64:
			args = append(args, strconv.FormatUint(tv, 10))
		default:
			args = append(args, "ERR")
		}
	}
	return args, nil
}
