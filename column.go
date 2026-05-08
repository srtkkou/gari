package gari

import (
	"errors"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
)

type (
	// 列
	column struct {
		table         *Table            // Pointer to table.
		name          string            // Table column name.
		fieldName     string            // Field name in struct.
		kind          kind              // Column type.
		size          int               // Size of column type.
		primary       bool              // Primary key.
		autoIncrement bool              // Flag to set AUTOINCREMENT.
		notNull       bool              // Flag to set NOT NULL. TODO Rename to nullable
		defaultValue  encoded           // Default value of column.
		rules         []validation.Rule // Validation rules.
		ddlCache      string            //Cached DDL SQL statement.
		//err           error             // Error
		//dbKind        sql.ColumnType    // Type on DB
	}
)

var (
	ErrColumnDDL = errors.New("gari.ErrColumnDDL")
)

// Create new column.
func newColumn(t *Table, name string) *column {
	c := &column{
		table:        t,
		name:         name,
		fieldName:    snakeToUpperCamelCase(name),
		kind:         "",
		size:         255,
		notNull:      false,
		defaultValue: []byte{},
		rules:        make([]validation.Rule, 0),
		ddlCache:     "",
	}
	return c
}

// 文字列化
func (c *column) String() string {
	ddl, _ := c.ddl()
	return ddl
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
func (c *column) ddl() (string, error) {
	// If present, return cached DDL SQL statement.
	if len(c.ddlCache) > 0 {
		return c.ddlCache, nil
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
	if c.primary {
		tokens = append(tokens, "PRIMARY KEY")
		if c.autoIncrement {
			tokens = append(tokens, "AUTOINCREMENT")
		}
	}
	// Null
	if c.notNull {
		tokens = append(tokens, "NOT NULL")
	}
	// Default
	if len(c.defaultValue) > 0 {
		switch c.kind {
		case kindString:
			ns, err := decodeNullString(c.defaultValue)
			if err != nil {
				err = errs.Wrap(ErrColumnDDL, errs.WithCause(err),
					errs.WithContext("columnName", c.name),
					errs.WithContext("default", c.defaultValue))
				return "", err
			}
			if ns.Valid {
				tokens = append(tokens, "DEFAULT")
				quote := c.table.gari.stringQuote
				tokens = append(tokens, quote+ns.String+quote)
			} else {
				tokens = append(tokens, "DEFAULT NULL")
			}
		case kindInt64:
			ni, err := decodeNullInt64(c.defaultValue)
			if err != nil {
				err = errs.Wrap(ErrColumnDDL, errs.WithCause(err),
					errs.WithContext("columnName", c.name),
					errs.WithContext("default", c.defaultValue))
				return "", err
			}
			if ni.Valid {
				tokens = append(tokens, "DEFAULT")
				tokens = append(tokens, strconv.FormatInt(ni.Int64, 10))
			} else {
				tokens = append(tokens, "DEFAULT NULL")
			}
		}
	}
	c.ddlCache = strings.Join(tokens, " ")
	return c.ddlCache, nil
}

// Add validation rule to column.
func (c *column) addRule(rule validation.Rule) {
	c.rules = append(c.rules, rule)
}
