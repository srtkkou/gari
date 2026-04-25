package gari

import (
	"fmt"
	"strconv"
	"strings"
)

type (
	// Builder to build DDL SQL.
	ddlBuilder struct {
		table *Table // Pointer to Table.
	}
)

func newDdlBuilder(t *Table) *ddlBuilder {
	db := ddlBuilder{
		table: t,
	}
	return &db
}

func (db *ddlBuilder) sql() (string, error) {
	strQuote := db.table.gari.stringQuote
	var b strings.Builder
	b.WriteString("CREATE TABLE ")
	b.WriteString("IF NOT EXISTS ")
	b.WriteString(db.table.Name())
	b.WriteString("(")
	for i, name := range db.table.columnNames {
		if i > 0 {
			b.WriteString(", ")
		}
		// Name
		b.WriteString(name)
		b.WriteString(" ")
		col := db.table.columns[name]
		// Type
		switch col.kind {
		case kindString:
			b.WriteString("TEXT")
		case kindTime:
			b.WriteString("TEXT")
		case kindInt64:
			b.WriteString("INTEGER")
		}
		// Null
		if col.notNull {
			b.WriteString(" NOT NULL")
		}
		// Default
		fmt.Printf("ColKind=%s default=%#x\n", col.kind, col.defaultValue)
		if len(col.defaultValue) > 0 {
			switch col.kind {
			case kindString:
				ns, err := decodeNullString(col.defaultValue)
				if err != nil {
					return "", err
				}
				if ns.Valid {
					b.WriteString(" DEFAULT ")
					b.WriteString(strQuote)
					b.WriteString(ns.String)
					b.WriteString(strQuote)
				} else {
					b.WriteString(" DEFAULT NULL")
				}
			case kindInt64:
				ni, err := decodeNullInt64(col.defaultValue)
				if err != nil {
					return "", err
				}
				if ni.Valid {
					b.WriteString(" DEFAULT ")
					b.WriteString(strconv.FormatInt(ni.Int64, 10))
				} else {
					b.WriteString(" DEFAULT NULL")
				}
			}
		}
	}
	b.WriteString(");")
	return b.String(), nil
}
