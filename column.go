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

// Column name with table.
func (c *column) fullName() string {
	var sb strings.Builder
	sb.WriteString(c.table.name)
	sb.WriteString(".")
	sb.WriteString(c.name)
	return sb.String()
}

// Quoted column name.
func (c *column) quotedName() string {
	return c.table.gari.quoteIdentifier(c.name)
}

// Quoted full column name.
func (c *column) quotedFullName() string {
	var sb strings.Builder
	sb.WriteString(c.table.quotedName())
	sb.WriteString(".")
	sb.WriteString(c.quotedName())
	return sb.String()
}

// Build DDL SQL.
func (c *column) query() string {
	// If present, return cached query.
	if len(c.queryCache) > 0 {
		return c.queryCache
	}
	tokens := make([]string, 0)
	// Name.
	tokens = append(tokens, c.name)
	// Type
	switch c.kind {
	case kindString:
		tokens = append(tokens, "TEXT")
	case kindTime:
		tokens = append(tokens, "DATETIME")
	case kindInt64:
		tokens = append(tokens, "INTEGER")
	}
	// Primary key and auto increment.
	if c.isPrimary {
		tokens = append(tokens, "PRIMARY", "KEY")
		if c.isAutoIncrement {
			tokens = append(tokens, "AUTOINCREMENT")
		}
	}
	// Null
	if !c.isNullable {
		tokens = append(tokens, "NOT", "NULL")
	}
	// Default
	if c.defaultExists {
		switch tv := c.defaultValue.(type) {
		case nil:
			tokens = append(tokens, "DEFAULT", "NULL")
		case string:
			token := c.gari().quoteString(tv)
			tokens = append(tokens, "DEFAULT", token)
		case int64:
			token := strconv.FormatInt(tv, 10)
			tokens = append(tokens, "DEFAULT", token)
		}
	}
	c.queryCache = strings.Join(tokens, " ")
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
