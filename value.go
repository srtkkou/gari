package gari

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/goark/errs"
)

type (
	// Value to store DB value.
	value struct {
		column *column // Pointer to Column
		blob   encoded // Encoded value

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
	v.gari().debugLog("value",
		slog.Any("input", value),
		slog.Any("blob", v.blob),
		slog.Any("err", err),
		slog.String("columnName", v.column.name),
		slog.String("kind", v.column.kind))
	return err
}

// Convert to SQL argument value.
func (c *value) arg() any {
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

func (v *value) gari() *Gari {
	return v.column.table.gari
}
