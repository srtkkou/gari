package gari

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type (
	// Value to store DB value.
	value struct {
		tableName  string // Table name.
		columnName string // Column name.
		fieldName  string // Golang struct field name.
		kind       kind   // Kind.
		dbType     string // DB type name.
		length     int64  // Length of DB column.
		precision  int64  // Decimal precision of DB column.
		scale      int64  // Decimal scale of DB column.
		isNullable bool   // Allow nil for DB column.
		raw        any    // Raw value.
	}
)

var (
	ErrValueScan = errors.New("gari.ErrValueScan")
)

// Create new Value struct by *column.
func newValueByColumn(col *column) *value {
	v := &value{}
	v.tableName = col.table.Name
	v.columnName = col.name
	v.fieldName = col.fieldName
	v.kind = col.kind
	v.dbType = col.dbType
	v.length = col.length
	v.precision = col.precision
	v.scale = col.scale
	v.isNullable = col.isNullable
	return v
}

// Create new Value struct by *column.
func newValueByColumnType(t *sql.ColumnType) *value {
	v := &value{}
	// Split name to table and column.
	tokens := strings.Split(t.Name(), ".")
	if len(tokens) == 1 {
		v.tableName = "unknown"
		v.columnName = tokens[0]
	} else {
		v.tableName = tokens[0]
		v.columnName = tokens[len(tokens)-1]
	}
	v.fieldName = snakeToUpperCamelCase(v.columnName)
	v.dbType = t.DatabaseTypeName()
	switch strings.ToUpper(v.dbType) {
	case "VARCHAR", "TEXT":
		v.kind = kindString
	case "DATE", "TIME", "DATETIME", "TIMESTAMP":
		v.kind = kindTime
	case "INTEGER", "INT":
		v.kind = kindInt64
	default:
		v.kind = kindInvalid
	}
	if length, ok := t.Length(); ok {
		v.length = length
	}
	if precision, scale, ok := t.DecimalSize(); ok {
		v.precision = precision
		v.scale = scale
	}
	if isNullable, ok := t.Nullable(); ok {
		v.isNullable = isNullable
	}
	return v
}

// Convert to JSON string.
func (v *value) String() string {
	var sb strings.Builder
	sb.WriteString(`{"type":"gari.value"`)
	sb.WriteString(`,"tableName":`)
	sb.WriteString(strconv.Quote(v.tableName))
	sb.WriteString(`,"columnName":`)
	sb.WriteString(strconv.Quote(v.columnName))
	sb.WriteString(`,"fieldName":`)
	sb.WriteString(strconv.Quote(v.fieldName))
	sb.WriteString(`,"kind":`)
	sb.WriteString(strconv.Quote(v.kind))
	sb.WriteString(`,"dbType":`)
	sb.WriteString(strconv.Quote(v.dbType))
	sb.WriteString(`,"length":`)
	sb.WriteString(strconv.FormatInt(v.length, 10))
	sb.WriteString(`,"precision":`)
	sb.WriteString(strconv.FormatInt(v.precision, 10))
	sb.WriteString(`,"scale":`)
	sb.WriteString(strconv.FormatInt(v.scale, 10))
	sb.WriteString(`,"isNullable":`)
	sb.WriteString(strconv.FormatBool(v.isNullable))
	sb.WriteString(`,"raw":`)
	raw := fmt.Sprintf("%v(%T)", v.raw, v.raw)
	sb.WriteString(strconv.Quote(raw))
	sb.WriteString(`}`)
	return sb.String()
}

// Scan DB value into value struct.
func (v *value) Scan(value any) (err error) {
	v.raw = value
	return err
}

// Convert to SQL argument value.
func (v *value) arg() any {
	return v.raw
}
