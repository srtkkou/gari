package gari

import (
	"errors"
	"slices"

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
	ErrDefaultValueNull   = errors.New("gari.ErrDefaultValueNull")
	ErrDefaultValueKind   = errors.New("gari.ErrDefaultValueKind")
	ErrDefaultValueEncode = errors.New("gari.ErrDefaultValueEncode")
	ErrDefaultValueSize   = errors.New("gari.ErrDefaultValueSize")
)

// Initialize columnBuilder.
func newColumnBuilder(col *column) *columnBuilder {
	return &columnBuilder{
		column: col,
	}
}

// Set field name.
func FieldName(name string) ColumnOption {
	return setColumnOption(func(b *columnBuilder) error {
		b.column.fieldName = name
		return nil
	})
}

// Set primary key flag.
func PrimaryKey(isAutoIncrement bool) ColumnOption {
	return setColumnOption(func(b *columnBuilder) error {
		b.column.primary = true
		b.column.autoIncrement = isAutoIncrement
		return nil
	})
}

// Set size of column.
func Size(size int) ColumnOption {
	return setColumnOption(func(b *columnBuilder) error {
		b.column.size = size
		return nil
	})
}

// Set NOT NULL for column.
func NotNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) error {
		b.column.notNull = true
		return nil
	})
}

// Set default value NULL.
func DefaultNull() ColumnOption {
	return setColumnOption(func(b *columnBuilder) error {
		if b.column.notNull {
			return errs.Wrap(ErrDefaultValueNull,
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("notNull", b.column.notNull))
		}
		b.column.defaultValue = msgpackNil()
		return nil
	})
}

// Set default bool value.
func DefaultBool(input bool) ColumnOption {
	return setColumnOption(func(b *columnBuilder) (err error) {
		if err = b.validateKind(kindBool); err != nil {
			return err
		}
		// Encode defalut bool value to msgpack.
		b.column.defaultValue, err = encodeBool(input)
		if err != nil {
			return errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", input))
		}
		return nil
	})
}

// Set default string value.
func DefaultString(str string) ColumnOption {
	return setColumnOption(func(b *columnBuilder) (err error) {
		if err = b.validateKind(kindString); err != nil {
			return err
		}
		// Validate field size.
		if b.column.size < len(str) {
			return errs.Wrap(ErrDefaultValueSize,
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", str),
				errs.WithContext("size", b.column.size))
		}
		// Encode default string value to msgpack.
		b.column.defaultValue, err = encodeString(str)
		if err != nil {
			return errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", str))
		}
		return nil
	})
}

// Set default int value.
func DefaultInt64(num int64) ColumnOption {
	return setColumnOption(func(b *columnBuilder) (err error) {
		if err = b.validateKind(kindInt64); err != nil {
			return err
		}
		// Encode default int64 value to msgpack.
		b.column.defaultValue, err = encodeInt64(num)
		if err != nil {
			return errs.Wrap(ErrDefaultValueEncode,
				errs.WithCause(err),
				errs.WithContext("columnName", b.column.name),
				errs.WithContext("input", num))
		}
		return nil
	})
}

// Wrapper func to set column option.
func setColumnOption(
	fn func(b *columnBuilder) error,
) ColumnOption {
	return func(b *columnBuilder) {
		// Do nothing if columnBuilder has error.
		if b.err != nil {
			return
		}
		// Set column option.
		if err := fn(b); err != nil {
			b.err = err
		}
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
