package gari

import (
	"database/sql"
	"math"
	"strings"
	"time"
)

type (
	// Data record.
	Record struct {
		values []*value          // Slice of values.
		m      map[string]*value // Map of values.
	}
)

func newRecord(values []*value) *Record {
	r := &Record{
		values: values,
		m:      make(map[string]*value, 0),
	}
	for _, value := range values {
		r.m[value.columnName] = value
		r.m[value.fieldName] = value
	}
	return r
}

func (r *Record) String() string {
	var sb strings.Builder
	sb.WriteString(`{"type":"gari.Record"`)
	sb.WriteString(`,"values":[`)
	for i, value := range r.values {
		if i > 0 {
			sb.WriteString(`,`)
		}
		sb.WriteString(value.String())
	}
	sb.WriteString(`]}`)
	return sb.String()
}

func (r *Record) ColumnNames() []string {
	names := make([]string, len(r.values))
	for i, value := range r.values {
		names[i] = value.columnName
	}
	return names
}

func (r *Record) FieldNames() []string {
	names := make([]string, len(r.values))
	for i, value := range r.values {
		names[i] = value.fieldName
	}
	return names
}

func (r *Record) SetNullString(name string, ptr *sql.NullString) {
	if value, ok := r.m[name]; ok {
		if ns, ok := value.raw.(sql.NullString); ok {
			*ptr = ns
		}
	}
}

func (r *Record) SetString(name string, ptr *string) {
	if value, ok := r.m[name]; ok {
		if str, ok := value.raw.(string); ok {
			*ptr = str
		}
	}
}

func (r *Record) SetNullTime(name string, ptr *sql.NullTime) {
	if value, ok := r.m[name]; ok {
		if nt, ok := value.raw.(sql.NullTime); ok {
			*ptr = nt
		}
	}
}

func (r *Record) SetTime(name string, ptr *time.Time) {
	if value, ok := r.m[name]; ok {
		if tm, ok := value.raw.(time.Time); ok {
			*ptr = tm
		}
	}
}

func (r *Record) SetNullInt64(name string, ptr *sql.NullInt64) {
	if value, ok := r.m[name]; ok {
		if ni, ok := value.raw.(sql.NullInt64); ok {
			*ptr = ni
		}
	}
}

func (r *Record) SetInt64(name string, ptr *int64) {
	if value, ok := r.m[name]; ok {
		switch tv := value.raw.(type) {
		case int:
			*ptr = int64(tv)
		case int8:
			*ptr = int64(tv)
		case int16:
			*ptr = int64(tv)
		case int32:
			*ptr = int64(tv)
		case int64:
			*ptr = tv
		}
	}
}

func (r *Record) SetInt(name string, ptr *int) {
	num := int64(0)
	r.SetInt64(name, &num)
	if math.MinInt <= num && num <= math.MaxInt {
		*ptr = int(num)
	}
}

func (r *Record) SetInt32(name string, ptr *int32) {
	num := int64(0)
	r.SetInt64(name, &num)
	if math.MinInt32 <= num && num <= math.MaxInt32 {
		*ptr = int32(num)
	}
}

func (r *Record) SetInt16(name string, ptr *int16) {
	num := int64(0)
	r.SetInt64(name, &num)
	if math.MinInt16 <= num && num <= math.MaxInt16 {
		*ptr = int16(num)
	}
}

func (r *Record) SetInt8(name string, ptr *int8) {
	num := int64(0)
	r.SetInt64(name, &num)
	if math.MinInt8 <= num && num <= math.MaxInt8 {
		*ptr = int8(num)
	}
}

func (r *Record) args() []any {
	args := make([]any, len(r.values))
	for i, value := range r.values {
		args[i] = value.arg()
	}
	return args
}

func (r *Record) valueByName(name string) (*value, bool) {
	v, ok := r.m[name]
	return v, ok
}
