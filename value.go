package gari

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goark/errs"
)

type (
	// Cell to store DB value.
	cell struct {
		column *column // Pointer to Column
		blob   encoded // Encoded value
	}
)

var (
	ErrCellScan = errors.New("gari.ErrCellScan")
)

// Create new Cell struct.
func newCell(col *column) *cell {
	c := cell{
		column: col,
		blob:   []byte{},
	}
	return &c
}

// String
func (c *cell) String() string {
	var b strings.Builder
	b.WriteString(`{"column":`)
	b.WriteString(c.column.String())
	b.WriteString(`}`)
	return b.String()
}

// Scan value into Cell struct.
func (c *cell) Scan(value any) (err error) {
	switch tv := value.(type) {
	case nil:
		c.blob = msgpackNil()
	case string:
		c.blob, err = encodeString(tv)
	case time.Time:
		c.blob, err = encodeTime(tv)
	case int:
		c.blob, err = encodeInt64(int64(tv))
	case int8:
		c.blob, err = encodeInt64(int64(tv))
	case int16:
		c.blob, err = encodeInt64(int64(tv))
	case int32:
		c.blob, err = encodeInt64(int64(tv))
	case int64:
		c.blob, err = encodeInt64(tv)
	default:
		err = errs.Wrap(ErrCellScan,
			errs.WithContext("input", value))
	}
	fmt.Printf("--Cell=%v(%T) col=%s blob=%#x err=%v\n",
		value, value, c.column.name, c.blob, err)
	return err
}

// Encode field name/value pair to msgpack.
func (c *cell) Encode() ([]byte, error) {
	blob, err := encodeString(c.column.fieldName)
	if err != nil {
		return nil, err
	}
	blob = append(blob, c.blob...)
	return blob, nil
}

// Convert to SQL argument value.
func (c *cell) arg() any {
	switch c.column.kind {
	case kindString:
		ns, err := decodeNullString(c.blob)
		if err != nil {
			return nil
		}
		if !ns.Valid {
			return nil
		}
		return ns.String
	case kindTime:
		nt, err := decodeNullTime(c.blob)
		if err != nil {
			return nil
		}
		if !nt.Valid {
			return nil
		}
		return nt.Time
	case kindInt64:
		ni, err := decodeNullInt64(c.blob)
		if err != nil {
			return nil
		}
		if !ni.Valid {
			return nil
		}
		return ni.Int64
	default:
		return nil
	}
}
