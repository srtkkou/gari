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
		column *Column // Pointer to Column
		value  []byte  // Encoded value
	}
)

var (
	ErrCellScan = errors.New("gari.ErrCellScan")
)

// Create new Cell struct.
func newCell(col *Column) *cell {
	c := cell{
		column: col,
		value:  []byte{},
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
		c.value = msgpackNil()
	case string:
		c.value, err = encodeString(tv)
	case time.Time:
		c.value, err = encodeTime(tv)
	case int:
		c.value, err = encodeInt64(int64(tv))
	case int8:
		c.value, err = encodeInt64(int64(tv))
	case int16:
		c.value, err = encodeInt64(int64(tv))
	case int32:
		c.value, err = encodeInt64(int64(tv))
	case int64:
		c.value, err = encodeInt64(tv)
	default:
		err = errs.Wrap(ErrCellScan,
			errs.WithContext("input", value))
	}
	fmt.Printf("--Cell=%v(%T) col=%s blob=%#x err=%v\n",
		value, value, c.column.name, c.value, err)
	return err
}

// Encode field name/value pair to msgpack.
func (c *cell) Encode() ([]byte, error) {
	blob, err := encodeString(c.column.fieldName)
	if err != nil {
		return nil, err
	}
	blob = append(blob, c.value...)
	return blob, nil
}
