package gari

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type (
	// Builder to build INSERT SQL.
	insertBuilder struct {
		ctx   context.Context // Context.
		table *Table          // Pointer to table.
	}
)

// Create new insertBuilder instance.
func newInsertBuilder(
	ctx context.Context, t *Table,
) *insertBuilder {
	b := insertBuilder{
		ctx:   ctx,
		table: t,
	}
	return &b
}

// Build SQL statement.
func (b *insertBuilder) stmt(ptrs []any) (string, error) {
	tokens := make([]string, 0)
	tokens = append(tokens, "INSERT INTO")
	tokens = append(tokens, b.table.name)
	tokens = append(tokens,
		"("+strings.Join(b.table.fieldNames, ", ")+")")
	tokens = append(tokens, "VALUES")

	blocks := make([]string, len(ptrs))
	for i, ptr := range ptrs {
		// Convert to msgpack.
		blob, err := msgpack.Marshal(ptr)
		if err != nil {
			return "", err
		}
		fmt.Printf("INSERT msgpack=%#x\n", blob)
		// Convert to map.
		m, err := decodeMap(blob)
		if err != nil {
			return "", err
		}
		for k, v := range m {
			fmt.Printf("key=%s,value=%v(%T)\n", k, v, v)
		}
		// Build token.
		values := make([]string, len(b.table.fieldNames))
		for j, fieldName := range b.table.fieldNames {
			//			col := b.table.columns[fieldName]
			v, ok := m[fieldName]
			if !ok {
				return "", errors.New("NO FIELD FOUND ON GIVEN STRUCT")
			}
			quote := b.table.gari.stringQuote
			switch tv := v.(type) {
			case nil:
				values[j] = "NULL"
			case string:
				values[j] = quote + tv + quote
			case time.Time:
				values[j] = quote + tv.Format("20060102 15:06:07.999999") + quote
			case int8:
				values[j] = strconv.FormatInt(int64(tv), 10)
			case int16:
				values[j] = strconv.FormatInt(int64(tv), 10)
			case int32:
				values[j] = strconv.FormatInt(int64(tv), 10)
			case int64:
				values[j] = strconv.FormatInt(tv, 10)
			case uint16:
				values[j] = strconv.FormatUint(uint64(tv), 10)
			case uint32:
				values[j] = strconv.FormatUint(uint64(tv), 10)
			case uint64:
				values[j] = strconv.FormatUint(tv, 10)
			default:
				values[j] = "ERR"
			}
		}
		// Build VALUES block.
		blocks[i] = "(" + strings.Join(values, ", ") + ")"
	}
	// Build insert SQL.
	token := strings.Join(blocks, ", ")
	tokens = append(tokens, token)
	stmt := strings.Join(tokens, " ") + ";"
	fmt.Printf("INSERT SQL=%s\n", stmt)
	return stmt, nil
}
