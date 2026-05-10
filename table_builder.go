package gari

import (
	"errors"

	"github.com/goark/errs"
)

type (
	// Table builder.
	TableBuilder struct {
		table *Table // Pointer to table struct.
		err   error  // Error.
	}
)

var (
	ErrNoPrimaryKey = errors.New("gari.ErrNoPrimaryKey")
)

// Add id.
func (b *TableBuilder) IdColumn() *TableBuilder {
	b.Int64Column("id", PrimaryKey(true))
	return b
}

// Add timestamps.
func (b *TableBuilder) TimeStampColumns() *TableBuilder {
	b.TimeColumn("created_at", DefaultNull())
	b.TimeColumn("updated_at", DefaultNull())
	// TODO: Add table callback to update them.
	return b
}

// Add soft delete column.
func (b *TableBuilder) SoftDeleteColumn() *TableBuilder {
	b.TimeColumn("deleted_at", DefaultNull())
	// TODO: Allow Destroy() func.
	return b
}

// Add string type column.
func (b *TableBuilder) StringColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	fn := func(col *column) {
		col.kind = kindString
		col.dbType = "TEXT" // TODO: Change to dialect.
	}
	return b.addColumn(name, fn, opts)
}

// Add time type column.
func (b *TableBuilder) TimeColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	fn := func(col *column) {
		col.kind = kindTime
		col.dbType = "DATETIME" // TODO: Change to dialect.
	}
	return b.addColumn(name, fn, opts)
}

// Add int64 type column.
func (b *TableBuilder) Int64Column(
	name string, opts ...ColumnOption,
) *TableBuilder {
	fn := func(col *column) {
		col.kind = kindInt64
		col.dbType = "INTEGER" // TODO: Change to dialect.
	}
	return b.addColumn(name, fn, opts)
}

// Add float64 type column.
func (b *TableBuilder) Float64Column(
	name string, opts ...ColumnOption,
) *TableBuilder {
	fn := func(col *column) {
		col.kind = kindFloat64
		col.dbType = "REAL" // TODO: Change to dialect.
	}
	return b.addColumn(name, fn, opts)
}

// Add blob byte column.
func (b *TableBuilder) BlobColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	fn := func(col *column) {
		col.kind = kindBlob
		col.dbType = "BLOB" // TODO: Change to dialect.
	}
	return b.addColumn(name, fn, opts)
}

// Define table.
func (b *TableBuilder) Define() (*Table, error) {
	if b.err != nil {
		return nil, b.err
	}
	// Check if primary key exists.
	if b.table.pkey == nil {
		err := errs.Wrap(ErrNoPrimaryKey)
		b.gari().errorLog(err.Error())
		return nil, err
	}
	return b.table, nil
}

// Define table or panic.
func (b *TableBuilder) MustDefine() *Table {
	t, err := b.Define()
	if err != nil {
		panic(err)
	}
	return t
}

// Add column with specified type.
func (b *TableBuilder) addColumn(
	name string, fn func(c *column), opts []ColumnOption,
) *TableBuilder {
	// Skip if tableBuilder has error.
	if b.err != nil {
		return b
	}
	// Create new column and set to columnBuilder.
	col := newColumn(b.table, name)
	cb := newColumnBuilder(col)
	fn(col)
	// Parse column options.
	for _, opt := range opts {
		opt(cb)
	}
	// Add columnBuilder error to tableBuilder.
	if cb.err != nil {
		b.err = cb.err
		return b
	}
	// Add column to table.
	b.table.addColumn(col)
	return b
}

func (b *TableBuilder) gari() *Gari {
	return b.table.gari
}
