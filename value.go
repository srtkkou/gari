package gari

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
)

type (
	// Value to store DB value.
	value struct {
		column    *column // Pointer to Column
		blob      encoded // Encoded value
		name      string
		typeName  string
		length    int64
		precision int64
		scale     int64
		nullable  bool
	}
)

var (
	ErrValueScan = errors.New("gari.ErrValueScan")
)

// Create new Value struct.
func newValue(col *column) *value {
	v := value{
		column: col,
		blob:   []byte{},
	}
	return &v
}

// String
func (c *value) String() string {
	var b strings.Builder
	b.WriteString(`{"blob":"`)
	b.WriteString(c.column.kind)
	b.WriteString(`", "column":`)
	b.WriteString(c.column.String())
	b.WriteString(`}`)
	return b.String()
}

// Scan value into Value struct.
func (v *value) Scan(value any) (err error) {
	switch tv := value.(type) {
	case nil:
		v.blob = msgpackNil()
	case string:
		v.blob, err = encodeString(tv)
	case time.Time:
		v.blob, err = encodeTime(tv)
	case int:
		v.blob, err = encodeInt64(int64(tv))
	case int8:
		v.blob, err = encodeInt64(int64(tv))
	case int16:
		v.blob, err = encodeInt64(int64(tv))
	case int32:
		v.blob, err = encodeInt64(int64(tv))
	case int64:
		v.blob, err = encodeInt64(tv)
	default:
		err = errs.Wrap(ErrValueScan,
			errs.WithContext("input", value))
	}
	/*
		v.gari().debugLog("value.Scan()",
			slog.Any("input", value),
			slog.Any("blob", v.blob),
			slog.Any("err", err),
			slog.String("columnName", v.column.name),
			slog.String("kind", v.column.kind))
	*/
	fmt.Printf("value.Scan() input=%v(%T) blob=%#x err=%v\n",
		value, value, v.blob, err)
	return err
}

// Convert to SQL argument value.
func (v *value) arg() (result any) {
	result = nil
	switch v.column.kind {
	case kindBool:
		nb, err := decodeNullBool(v.blob)
		if err == nil && nb.Valid {
			result = nb.Bool
		}
	case kindString:
		ns, err := decodeNullString(v.blob)
		if err == nil && ns.Valid {
			result = ns.String
		}
	case kindTime:
		nt, err := decodeNullTime(v.blob)
		if err == nil && nt.Valid {
			result = nt.Time
		}
	case kindInt64:
		ni, err := decodeNullInt64(v.blob)
		if err == nil && ni.Valid {
			result = ni.Int64
		}
	}
	v.gari().debugLog("value.arg()",
		slog.Any("blob", v.blob),
		slog.Any("result", result),
		slog.String("columnName", v.column.name),
		slog.String("kind", v.column.kind))
	return result
}

func (v *value) gari() *Gari {
	return v.column.table.gari
}
