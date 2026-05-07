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
	ErrDefaultValueEncode  = errors.New("gari.ErrDefaultValueEncode")
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
		b.column.primary = true
		b.column.autoIncrement = isAutoIncrement
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
func Size(size int) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.size = size
	})
}

// Set NOT NULL for column.
func NotNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		b.column.notNull = true
		b.column.addRule(validation.Required)
	})
}

// Set default value NULL.
func DefaultNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if b.column.notNull {
			b.err = errs.Wrap(ErrDefaultValueNull,
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("notNull", b.column.notNull))
			return
		}
		b.column.defaultValue = msgpackNil()
	})
}

// Set default bool value.
func DefaultBool(input bool) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if err := b.validateKind(kindBool); err != nil {
			b.err = err
			return
		}
		// Encode defalut bool value to msgpack.
		blob, err := encodeBool(input)
		if err != nil {
			b.err = errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", input))
			return
		}
		b.column.defaultValue = blob
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
		if b.column.size < len(str) {
			b.err = errs.Wrap(ErrDefaultValueSize,
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", str),
				errs.WithContext("size", b.column.size))
			return
		}
		// Encode default string value to msgpack.
		blob, err := encodeString(str)
		if err != nil {
			b.err = errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", str))
			return
		}
		b.column.defaultValue = blob
	})
}

// Set default int value.
func DefaultInt64(num int64) ColumnOption {
	return setColumnOption(func(b *columnBuilder) {
		if err := b.validateKind(kindInt64); err != nil {
			b.err = err
			return
		}
		// Encode default int64 value to msgpack.
		blob, err := encodeInt64(num)
		if err != nil {
			b.err = errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", num))
			return
		}
		b.column.defaultValue = blob
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
