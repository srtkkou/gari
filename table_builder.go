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
	name string, fn func(col *Column),
) *TableBuilder {
	return b.addColumnWithKind(name, kindBool, fn)
}

// Add string type column.
func (b *TableBuilder) AddStringColumn(
	name string, fn func(col *Column),
) *TableBuilder {
	return b.addColumnWithKind(name, kindString, fn)
}

// Add time type column.
func (b *TableBuilder) AddTimeColumn(
	name string, fn func(col *Column),
) *TableBuilder {
	return b.addColumnWithKind(name, kindTime, fn)
}

// Add int64 type column.
func (b *TableBuilder) AddInt64Column(
	name string, fn func(col *Column),
) *TableBuilder {
	return b.addColumnWithKind(name, kindInt64, fn)
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
	name string, k kind, fn func(col *Column),
) *TableBuilder {
	col := newColumn(name)
	col.kind = k
	fn(col)
	b.table.addColumn(col)
	return b
}
