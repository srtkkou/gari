package gari

import (
	"database/sql"
	"math"
	"time"
)

type (
	// Data record.
	Record struct {
		cells []*cell          // Slice of cells.
		m     map[string]*cell // Map of cells.
	}
)

func newRecord(cells []*cell) *Record {
	r := &Record{
		cells: cells,
		m:     make(map[string]*cell, 0),
	}
	for _, cell := range cells {
		col := cell.column
		r.m[col.name] = cell
		r.m[col.fullName()] = cell
		r.m[col.fieldName] = cell
	}
	return r
}

func (r *Record) ColumnNames() []string {
	names := make([]string, len(r.cells))
	for i, cell := range r.cells {
		names[i] = cell.column.name
	}
	return names
}

func (r *Record) FullColumnNames() []string {
	names := make([]string, len(r.cells))
	for i, cell := range r.cells {
		names[i] = cell.column.fullName()
	}
	return names
}

func (r *Record) FieldNames() []string {
	names := make([]string, len(r.cells))
	for i, cell := range r.cells {
		names[i] = cell.column.fieldName
	}
	return names
}

func (r *Record) SetNullString(name string, ptr *sql.NullString) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindString && !col.notNull {
			if ns, err := decodeNullString(cell.blob); err == nil {
				*ptr = ns
			}
		}
	}
}

func (r *Record) SetString(name string, ptr *string) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindString && col.notNull {
			if ns, err := decodeNullString(cell.blob); err == nil {
				if ns.Valid {
					*ptr = ns.String
				}
			}
		}
	}
}

func (r *Record) SetNullTime(name string, ptr *sql.NullTime) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindTime && !col.notNull {
			if nt, err := decodeNullTime(cell.blob); err == nil {
				*ptr = nt
			}
		}
	}
}

func (r *Record) SetTime(name string, ptr *time.Time) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindTime && col.notNull {
			if nt, err := decodeNullTime(cell.blob); err == nil {
				if nt.Valid {
					*ptr = nt.Time
				}
			}
		}
	}
}

func (r *Record) SetNullInt64(name string, ptr *sql.NullInt64) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindInt64 && !col.notNull {
			if ni, err := decodeNullInt64(cell.blob); err == nil {
				*ptr = ni
			}
		}
	}
}

func (r *Record) SetInt64(name string, ptr *int64) {
	if cell, ok := r.m[name]; ok {
		col := cell.column
		if col.kind == kindInt64 && col.notNull {
			if ni, err := decodeNullInt64(cell.blob); err == nil {
				if ni.Valid {
					*ptr = ni.Int64
				}
			}
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
	args := make([]any, len(r.cells))
	for i, cell := range r.cells {
		args[i] = cell.arg()
	}
	return args
}
