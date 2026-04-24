package gari

import ()

type (
	// Table builder.
	TableBuilder struct {
		table *Table // Pointer to table struct.
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
	for _, name := range b.table.columnNames {
		col := b.table.columns[name]
		if col.err != nil {
			return nil, col.err
		}
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
	col := newColumn(name)
	col.kind = k
	for _, opt := range opts {
		opt(col)
	}
	b.table.addColumn(col)
	return b
}
