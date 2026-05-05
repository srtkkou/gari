package gari

import ()

type (
	// Table builder.
	TableBuilder struct {
		table *Table // Pointer to table struct.
		err   error  // Error.
	}
)

// Add bool type column.
func (b *TableBuilder) AddBoolColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	return b.addColumnWithKind(name, kindBool, opts)
}

// Add string type column.
func (b *TableBuilder) AddStringColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	return b.addColumnWithKind(name, kindString, opts)
}

// Add time type column.
func (b *TableBuilder) AddTimeColumn(
	name string, opts ...ColumnOption,
) *TableBuilder {
	return b.addColumnWithKind(name, kindTime, opts)
}

// Add int64 type column.
func (b *TableBuilder) AddInt64Column(
	name string, opts ...ColumnOption,
) *TableBuilder {
	return b.addColumnWithKind(name, kindInt64, opts)
}

// Define table.
func (b *TableBuilder) Define() (*Table, error) {
	if b.err != nil {
		return nil, b.err
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
func (b *TableBuilder) addColumnWithKind(
	name string, k kind, opts []ColumnOption,
) *TableBuilder {
	// Skip if tableBuilder has error.
	if b.err != nil {
		return b
	}
	// Create new column and set to columnBuilder.
	col := newColumn(name)
	cb := newColumnBuilder(col)
	col.kind = k
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
