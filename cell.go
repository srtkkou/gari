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

// Attr name and value.
func (c Cell) MsgpackBytes() ([]byte, error) {
	blob, err := encodeNullString(c.column.attrName)
	if err != nil {
		return []byte{}, err
	}
	blob = append(blob, c.value...)
	return blob, nil
}

// Scan value into Cell struct.
func (c *Cell) Scan(value any) error {
	switch c.column.kind {
	case kindString:
		return c.scanNullString(value)
	case kindTime:
		return c.scanNullTime(value)
	case kindInt64:
		return c.scanNullInt64(value)
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

// Scan sql.NullInt64 type.
func (c *Cell) scanNullInt64(value any) error {
	ni := sql.NullInt64{}
	err := ni.Scan(value)
	if err != nil {
		err = errs.Wrap(ErrScan, errs.WithCause(err),
			errs.WithContext("value", value))
		return err
	}
	fmt.Printf("NullInt:Valid=%t,Int64=%v\n", ni.Valid, ni.Int64)
	// Store value
	blob, err := encodeNullInt64(ni)
	if err != nil {
		return err
	}
	c.value = blob
	return nil
}
