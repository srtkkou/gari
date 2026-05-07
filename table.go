package gari

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/goark/errs"
)

type (
	// Table
	Table struct {
		gari          *Gari              // Pointer to gari config.
		name          string             // Table name.
		columns       []*column          // Slice of pointer to columns.
		columnMap     map[string]*column // Map of columns.
		pkey          *column            // Pointer to PRIMARY KEY column.
		insertBuilder *insertBuilder     // Builder to execute INSERT SQL.
		updateBuilder *updateBuilder     // Builder to execute UPDATE SQL.
		deleteBuilder *deleteBuilder     // Builder to execute DELETE SQL.
	}
)

var (
	ErrTableMigrate = errors.New("gari.ErrTableMigrate")
)

// Create new table.
func newTable(g *Gari, name string) *Table {
	t := Table{
		gari:      g,
		name:      name,
		columns:   make([]*column, 0),
		columnMap: make(map[string]*column, 0),
	}
	return &t
}

// Table name.
func (t *Table) Name() string {
	return t.name
}

// Column names.
func (t *Table) ColumnNames() []string {
	names := make([]string, len(t.columns))
	for i, col := range t.columns {
		names[i] = col.name
	}
	return names
}

// Column names with table name.
func (t *Table) FullColumnNames() []string {
	names := make([]string, len(t.columns))
	for i, col := range t.columns {
		names[i] = fmt.Sprintf("%s.%s", t.name, col.name)
	}
	return names
}

// Struct field names.
func (t *Table) FieldNames() []string {
	names := make([]string, len(t.columns))
	for i, col := range t.columns {
		names[i] = col.fieldName
	}
	return names
}

// Build SELECT SQL statement.
func (t *Table) Select() *selectBuilder {
	return newSelectBuilder(t)
}

// Execute INSERT SQL statement.
func (t *Table) Insert() *insertBuilder {
	if t.insertBuilder != nil {
		return t.insertBuilder
	}
	t.insertBuilder = newInsertBuilder(t)
	t.insertBuilder.prepareStmt()
	return t.insertBuilder
}

// Execute UPDATE SQL statement.
func (t *Table) Update() *updateBuilder {
	if t.updateBuilder != nil {
		return t.updateBuilder
	}
	t.updateBuilder = newUpdateBuilder(t)
	t.updateBuilder.prepareStmt()
	return t.updateBuilder
}

// Execute DELETE SQL statement.
func (t *Table) Delete() *deleteBuilder {
	if t.deleteBuilder != nil {
		return t.deleteBuilder
	}
	t.deleteBuilder = newDeleteBuilder(t)
	t.deleteBuilder.prepareStmt()
	return t.deleteBuilder
}

// Execute DDL SQL to migrate table.
func (t *Table) Migrate(ctx context.Context) error {
	// Build DDL SQL.
	ddl, err := newDdlBuilder(t).build()
	fmt.Printf("MIGRATE(err=%v)\nSQL=%s\n", err, ddl)
	// Execute query.
	db := t.gari.db
	startedAt := time.Now()
	_, err = db.ExecContext(ctx, ddl)
	if err != nil {
		err = errs.Wrap(ErrTableMigrate, errs.WithCause(err),
			errs.WithContext("query", ddl),
			errs.WithContext("duration", time.Since(startedAt)))
		t.gari.errorLog(err.Error())
		return err
	}
	t.gari.infoLog("Table.Migrate()",
		slog.String("query", ddl),
		slog.Duration("duration", time.Since(startedAt)))
	return nil
}

// Close prepared statements.
func (t *Table) closeTable() (err error) {
	if t.insertBuilder != nil {
		t.insertBuilder.closePreparedStmt()
	}
	if t.updateBuilder != nil {
		t.updateBuilder.closePreparedStmt()
	}
	if t.deleteBuilder != nil {
		t.deleteBuilder.closePreparedStmt()
	}
	return nil
}

/*
// sqlmock用の値をdriver.Valueのスライスとして取得する。
func (t *Table) DriverValues() []driver.Value {
	dbMap := t.toDbMap()
	values := make([]driver.Value, 0, len(dbMap))
	for _, column := range t.schema.columnMap {
		v := dbMap[column]
		if value, ok := v.(driver.Value); ok {
			values = append(values, value)
		}
	}
	return values
}

// テストデータの読み込み
func (t *Table) LoadTestdata(xid string) error {
	m, err := LoadTestdata(t.TableName(), xid)
	if err != nil {
		return err
	}
	return t.SetFromDbMap(m, t.TableName())
}
*/

// Add column to table
func (t *Table) addColumn(col *column) {
	t.columns = append(t.columns, col)
	// Add column pointer to map.
	t.columnMap[col.name] = col
	t.columnMap[col.fieldName] = col
}

// Get column by column name or field name.
func (t *Table) column(name string) *column {
	return t.columnMap[name]
}
