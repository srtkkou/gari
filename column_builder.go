package gari

import (
	"errors"
	"slices"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
)

type (
	// Options to define column.
	ColumnOption func(*columnBuilder)
	// Column builder
	columnBuilder struct {
		column *column // Pointer to column.
		err    error   // Error
	}
)

var (
	ErrMultiplePrimaryKeys = errors.New("gari.ErrMultiplePrimaryKeys")
	ErrDefaultValueNull    = errors.New("gari.ErrDefaultValueNull")
	ErrDefaultValueKind    = errors.New("gari.ErrDefaultValueKind")
	ErrDefaultValueSize    = errors.New("gari.ErrDefaultValueSize")
)

// Initialize columnBuilder.
func newColumnBuilder(col *column) *columnBuilder {
	return &columnBuilder{
		column: col,
	}
}

// Set field name.
func FieldName(name string) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.fieldName = name
	})
}

// Set primary key flag.
func PrimaryKey(isAutoIncrement bool) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.isPrimary = true
		b.column.isAutoIncrement = isAutoIncrement
		// Check if primary key is already set to table.
		if b.column.table.pkey != nil {
			b.err = errs.Wrap(ErrMultiplePrimaryKeys)
			return
		}
		// Store pointer to this column.
		b.column.table.pkey = b.column
	})
}

// Set size of column.
func Length(length int64) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.length = length
	})
}

// Set NOT NULL for column.
func NotNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.isNullable = false
		b.column.addRule(validation.Required)
	})
}

// Set default value NULL.
func DefaultNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if !b.column.isNullable {
			b.err = errs.Wrap(ErrDefaultValueNull,
				errs.WithContext("column", b.column.String()))
			return
		}
		b.column.defaultExists = true
		b.column.defaultValue = nil
	})
}

// Set default string value.
func DefaultString(str string) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if err := b.validateKind(kindString); err != nil {
			b.err = err
			return
		}
		// Validate field size.
		if b.column.length < int64(len(str)) {
			b.err = errs.Wrap(ErrDefaultValueSize,
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", str),
				errs.WithContext("length", b.column.length))
			return
		}
		b.column.defaultExists = true
		b.column.defaultValue = str
	})
}

// Set default int value.
func DefaultInt64(num int64) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if err := b.validateKind(kindInt64); err != nil {
			b.err = err
			return
		}
		b.column.defaultExists = true
		b.column.defaultValue = num
	})
}

// Wrapper func to set column option.
func setColumnOption(
	fn func(b *columnBuilder),
) ColumnOption {
	return func(b *columnBuilder) {
		// Do nothing if columnBuilder has error.
		if b.err != nil {
			return
		}
		// Set column option.
		fn(b)
	}
}

// Validate column kind.
func (b *columnBuilder) validateKind(kinds ...string) error {
	if !slices.Contains(kinds, b.column.kind) {
		return errs.Wrap(ErrDefaultValueKind,
			errs.WithContext("columnName", b.column.name),
			errs.WithContext("kind", b.column.kind))
	}
	return nil
}
