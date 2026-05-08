package gari

import (
	"database/sql"
	"math"
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

func (r *Record) ColumnNames() []string {
	names := make([]string, len(r.values))
	for i, value := range r.values {
		names[i] = value.columnName
	}
	return names
}

/*
func (r *Record) FullColumnNames() []string {
	names := make([]string, len(r.values))
	for i, value := range r.values {
		names[i] = value.column.fullName()
	}
	return names
}
*/

func (r *Record) FieldNames() []string {
	names := make([]string, len(r.values))
	for i, value := range r.values {
		names[i] = value.fieldName
	}
	return names
}

func (r *Record) SetNullString(name string, ptr *sql.NullString) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindString && value.nullable {
		if ns, ok := value.raw.(sql.NullString); ok {
			*ptr = ns
		}
		/*
			if ns, err := decodeNullString(value.blob); err == nil {
				*ptr = ns
			}
		*/
		//}
	}
}

func (r *Record) SetString(name string, ptr *string) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindString && !value.nullable {
		if str, ok := value.raw.(string); ok {
			*ptr = str
		}
		/*
			if ns, err := decodeNullString(value.blob); err == nil {
				if ns.Valid {
					*ptr = ns.String
				}
			}
		*/
		//}
	}
}

func (r *Record) SetNullTime(name string, ptr *sql.NullTime) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindTime && value.nullable {
		if nt, ok := value.raw.(sql.NullTime); ok {
			*ptr = nt
		}
		/*
			if nt, err := decodeNullTime(value.blob); err == nil {
				*ptr = nt
			}
		*/
		//}
	}
}

func (r *Record) SetTime(name string, ptr *time.Time) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindTime && !value.nullable {
		if tm, ok := value.raw.(time.Time); ok {
			*ptr = tm
		}
		/*
			if nt, err := decodeNullTime(value.blob); err == nil {
				if nt.Valid {
					*ptr = nt.Time
				}
			}
		*/
		//}
	}
}

func (r *Record) SetNullInt64(name string, ptr *sql.NullInt64) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindInt64 && value.nullable {
		if ni, ok := value.raw.(sql.NullInt64); ok {
			*ptr = ni
		}
		/*
			if ni, err := decodeNullInt64(value.blob); err == nil {
				*ptr = ni
			}
		*/
		//}
	}
}

func (r *Record) SetInt64(name string, ptr *int64) {
	if value, ok := r.m[name]; ok {
		//if value.kind == kindInt64 && !value.nullable {
		if num, ok := value.raw.(int64); ok {
			*ptr = num
		}
		/*
			if ni, err := decodeNullInt64(value.blob); err == nil {
				if ni.Valid {
					*ptr = ni.Int64
				}
			}
		*/
		//}
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
