package gari

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/goark/errs"
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
	// Scan error
	ErrScan = errors.New("orm.ErrScan")
)

// Create new Cell struct.
func newCell(col *Column) *Cell {
	c := Cell{
		column: col,
		value:  []byte{},
	}
	return &c
}

// String
func (c Cell) String() string {
	var b strings.Builder
	b.WriteString(`{"column":`)
	b.WriteString(c.column.String())
	b.WriteString(`}`)
	return b.String()
}

// Scan value into Cell struct.
func (c *Cell) Scan(value any) error {
	switch c.column.kind {
	case kindString:
		return c.scanNullString(value)
	case kindTime:
		return c.scanNullTime(value)
	default:
		return ErrCellKind
	}
}

// Scan sql.NullString type.
func (c *Cell) scanNullString(value any) error {
	ns := sql.NullString{}
	err := ns.Scan(value)
	if err != nil {
		err = errs.Wrap(ErrScan, errs.WithCause(err),
			errs.WithContext("value", value))
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

// Scan sql.NullTime type.
func (c *Cell) scanNullTime(value any) error {
	nt := sql.NullTime{}
	err := nt.Scan(value)
	if err != nil {
		err = errs.Wrap(ErrScan, errs.WithCause(err),
			errs.WithContext("value", value))
		return err
	}
	fmt.Printf("NullTime:Valid=%t,Time=%v\n", nt.Valid, nt.Time)
	// Store value
	blob, err := encodeNullTime(nt)
	if err != nil {
		return err
	}
	c.value = blob
	return nil
}
