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
		gari           *Gari              // Pointer to gari config.
		name           string             // Table name.
		columns        []*column          // Slice of pointer to columns.
		columnMap      map[string]*column // Map of columns.
		pkey           *column            // Pointer to PRIMARY KEY column.
		insertExecutor *insertExecutor    // Executor to execute INSERT SQL.
		updateExecutor *updateExecutor    // Executor to execute UPDATE SQL.
		deleteExecutor *deleteExecutor    // Executor to execute DELETE SQL.

		beforeInsertHooks []HookFunc
		afterInsertHooks  []HookFunc
		beforeUpdateHooks []HookFunc
		afterUpdateHooks  []HookFunc
		beforeDeleteHooks []HookFunc
		afterDeleteHooks  []HookFunc
	}
)

var (
	ErrTableMigrate = errors.New("gari.ErrTableMigrate")
)

// Create new table.
func newTable(g *Gari, name string) *Table {
	t := &Table{
		gari:      g,
		name:      name,
		columns:   make([]*column, 0),
		columnMap: make(map[string]*column, 0),
	}
	return t
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

// Execute INSERT SQL statement.
func (t *Table) Insert() *insertExecutor {
	if t.insertExecutor != nil {
		return t.insertExecutor
	}
	t.insertExecutor = newInsertExecutor(t)
	t.insertExecutor.prepareStmt()
	return t.insertExecutor
}

// Execute UPDATE SQL statement.
func (t *Table) Update() *updateExecutor {
	if t.updateExecutor != nil {
		return t.updateExecutor
	}
	t.updateExecutor = newUpdateExecutor(t)
	t.updateExecutor.prepareStmt()
	return t.updateExecutor
}

// Execute DELETE SQL statement.
func (t *Table) Delete() *deleteExecutor {
	if t.deleteExecutor != nil {
		return t.deleteExecutor
	}
	t.deleteExecutor = newDeleteExecutor(t)
	t.deleteExecutor.prepareStmt()
	return t.deleteExecutor
}

// Execute DDL SQL to migrate table.
func (t *Table) Migrate(ctx context.Context) error {
	// Build DDL SQL.
	query, err := newDdlBuilder(t).build()
	if err != nil {
		err = errs.Wrap(ErrTableMigrate, errs.WithCause(err))
		t.gari.errorLog(err.Error())
		return err
	}
	// Execute query.
	db := t.gari.db
	startedAt := time.Now()
	_, err = db.ExecContext(ctx, query)
	if err != nil {
		err = errs.Wrap(ErrTableMigrate, errs.WithCause(err),
			errs.WithContext("query", query),
			errs.WithContext("duration", time.Since(startedAt)))
		t.gari.errorLog(err.Error())
		return err
	}
	t.gari.infoLog("Table.Migrate()",
		slog.String("query", query),
		slog.Duration("duration", time.Since(startedAt)))
	return nil
}

func (t *Table) addHook(ht HookType, hf HookFunc) bool {
	switch ht {
	case BeforeInsert:
		t.beforeInsertHooks = append(t.beforeInsertHooks, hf)
	case AfterInsert:
		t.afterInsertHooks = append(t.afterInsertHooks, hf)
	case BeforeUpdate:
		t.beforeUpdateHooks = append(t.beforeUpdateHooks, hf)
	case AfterUpdate:
		t.afterUpdateHooks = append(t.afterUpdateHooks, hf)
	case BeforeDelete:
		t.beforeDeleteHooks = append(t.beforeDeleteHooks, hf)
	case AfterDelete:
		t.afterDeleteHooks = append(t.afterDeleteHooks, hf)
	}
	return true
}

// Close prepared statements.
func (t *Table) closeTable() (err error) {
	if t.insertExecutor != nil {
		t.insertExecutor.closePreparedStmt()
	}
	if t.updateExecutor != nil {
		t.updateExecutor.closePreparedStmt()
	}
	if t.deleteExecutor != nil {
		t.deleteExecutor.closePreparedStmt()
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
