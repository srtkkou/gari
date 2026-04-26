package gari

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build UPDATE SQL.
	updateBuilder struct {
		ctx        context.Context    // Context.
		table      *Table             // Pointer to table.
		maps       [](map[string]any) // Maps of structs.
		fieldNames []string           // Sorted field names of struct.
	}
)

var (
	ErrUpdateBuilderParse = errors.New("gari.ErrUpdateBuilderParse")
)

// Create new updateBuilder instance.
func newUpdateBuilder(
	ctx context.Context, t *Table,
) *updateBuilder {
	b := updateBuilder{
		ctx:   ctx,
		table: t,
		maps:  make([]map[string]any),
	}
	return &b
}

// Parse ptrs.
func (b *updateBuilder) parse(ptrs []any) error {
	// Convert to msgpack.
	blob, err := msgpack.Marshal(ptr)
	if err != nil {
		return errs.Wrap(ErrUpdateBuilderParse, errs.WithCause(err))
	}
	// Unmarshal msgpack.
	err := msgpack.Unmarshal(blob, &b.maps)
	if err != nil {
		return errs.Wrap(ErrUpdateBuilderParse, errs.WithCause(err),
			errs.WithContext("msgpack", blob))
	}
	// Check size of maps.
	if len(maps) == 0 {
		return errs.Wrap(ErrUpdateBuilderParse,
			errs.WithContext("maps", maps))
	}
	// Prepare sorted field names.
	b.fieldNames := make([]string, len(b.maps), 0)
	for k := range b.maps[0] {
		if k != "Id" {
			b.fieldNames = append(b.fieldNames, k)
		}
	}
	sort.Strings(b.fieldNames)
	b.fieldNames = append(b.fieldNames, "Id")
	return nil
}

// Build SQL statement.
func (b *updateBuilder) query() (string, error) {
	tokens := make([]string, 0)
	tokens = append(tokens, "UPDATE")
	tokens = append(tokens, b.table.name)
	tokens = append(tokens, "SET")
	// Build token.
	values := make([]string, len(b.maps), 0)
	for _, fieldName := range b.fieldNames {
		col, ok := b.table.columns[fieldName]
		if !ok {
			return "", fmt.Errorf("%s is not a table field", k)
		}
		values = append(values, col.name+" = ?")
	}
	tokens = append(tokens, strings.Join(values, ", "))
	tokens = append(tokens, "WHERE ID = ?")
	stmt := strings.Join(tokens, " ") + ";"
	fmt.Printf("UPDATE SQL=%s\n", stmt)
	return stmt, nil
}

// Build argument of ptr.
func (b *updateBuilder) args() ([]any, error) {
	quote := b.table.gari.stringQuote
	args := make([]any, len(b.maps), 0)
	for k, v := range b.fieldNames {
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
