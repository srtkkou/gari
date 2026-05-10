package gari

import (
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// 列
	column struct {
		table           *Table            // Pointer to table.
		name            string            // Column name.
		fieldName       string            // Golang struct Field name.
		kind            kind              // Kind.
		dbType          string            // DB type name.
		length          int64             // Length of DB column.
		precision       int64             // Decimal precision of DB column.
		scale           int64             // Decimal precision of DB column.
		isNullable      bool              // Allow nil for DB column.
		defaultExists   bool              // Default value is set.
		defaultValue    any               // Default raw value.
		isPrimary       bool              // Primary key.
		isAutoIncrement bool              // Set auto-increment for isPrimary key.
		rules           []validation.Rule // Validation rules.
		queryCache      string            //Cached DDL SQL statement.
	}
)

// Create new column.
func newColumn(t *Table, name string) *column {
	c := &column{
		table:      t,
		name:       name,
		fieldName:  snakeToUpperCamelCase(name),
		length:     255,
		isNullable: true,
		rules:      make([]validation.Rule, 0),
	}
	return c
}

// Convert to JSON string.
func (c *column) String() string {
	var sb strings.Builder
	sb.WriteString(`{"type":"gari.column"`)
	sb.WriteString(`,"name":`)
	sb.WriteString(strconv.Quote(c.name))
	sb.WriteString(`,"fieldName":`)
	sb.WriteString(strconv.Quote(c.fieldName))
	sb.WriteString(`,"kind":`)
	sb.WriteString(strconv.Quote(c.kind))
	sb.WriteString(`,"query":`)
	sb.WriteString(strconv.Quote(c.query()))
	sb.WriteString(`}`)
	return sb.String()
}

// Build DDL SQL.
func (c *column) query() string {
	// If present, return cached query.
	if len(c.queryCache) > 0 {
		return c.queryCache
	}
	dialect := c.gari().dialect
	var sb strings.Builder
	// Name.
	sb.WriteString(c.name)
	sb.WriteString(` `)
	// Type
	switch c.kind {
	case kindString:
		sb.WriteString(`TEXT`)
	case kindTime:
		sb.WriteString(`DATETIME`)
	case kindInt64:
		sb.WriteString(`INTEGER`)
	}
	// Primary key and auto increment.
	if c.isPrimary {
		sb.WriteString(` PRIMARY KEY`)
		if c.isAutoIncrement {
			sb.WriteString(` AUTOINCREMENT`)
		}
	}
	// Null
	if !c.isNullable {
		sb.WriteString(` NOT NULL`)
	}
	// Default
	if c.defaultExists {
		switch tv := c.defaultValue.(type) {
		case nil:
			sb.WriteString(` DEFAULT NULL`)
		case string:
			sb.WriteString(` DEFAULT `)
			dialect.QuoteString(&sb, tv)
		case int64:
			sb.WriteString(` DEFAULT `)
			sb.WriteString(strconv.FormatInt(tv, 10))
		}
	}
	c.queryCache = sb.String()
	return c.queryCache
}

// Add validation rule to column.
func (c *column) addRule(rule validation.Rule) {
	c.rules = append(c.rules, rule)
}

// Shorthand func for *Gari.
func (c *column) gari() *Gari {
	return c.table.gari
}
