package gari

import (
	"database/sql"
	"errors"
	"fmt"
	//	"slices"
	//	"strconv"
	"strings"
	//	"github.com/goark/errs"
)

type (
	// Cell
	// Box to store DB value.
	Cell struct {
		column *Column // Pointer to Column
		value  []byte  // Encoded value
	}
)

var (
	// Cell kind error
	ErrCellKind = errors.New("orm.ErrCellKind")
)

// Create new Cell struct.
func newCell(col *Column) *Cell {
	c := Cell{
		column: col,
		value:  []byte{},
	}
	return &c
}

/*
// 列の新規作成
func newColumn(name string) *Column {
	return &Column{
		name:         name,
		attrName:     snakeToUpperCamelCase(name),
		jsonKey:      snakeToLowerCamelCase(name),
		kind:         "",
		size:         -1,
		allowNull:    false,
		value:        []byte{},
		defaultValue: []byte{},
		rules:        make([]validation.Rule, 0),
	}
}
*/

// 文字列化
func (c Cell) String() string {
	var b strings.Builder
	b.WriteString(`{"column":`)
	b.WriteString(c.column.String())
	b.WriteString(`}`)
	return b.String()
}

func (c *Cell) Scan(value any) error {
	switch c.column.kind {
	case kindString:
		return c.scanNullString(value)
	default:
		return ErrCellKind
	}
}

func (c *Cell) scanNullString(value any) error {
	ns := sql.NullString{}
	err := ns.Scan(value)
	if err != nil {
		fmt.Printf("ScanError=%v\n", err)
		return err
	}
	fmt.Printf("NullString:Valid=%t,String=%s\n", ns.Valid, ns.String)
	// Store value
	blob, err := encodeNullString(ns)
	if err != nil {
		return err
	}
	c.value = blob
	return nil
}
