package gari

import (
	"context"
	"database/sql"
	"fmt"
)

type (
	// Table
	Table struct {
		gari        *Gari              // Pointer to gari config.
		name        string             // Table name.
		columnNames []string           // Column names.
		fieldNames  []string           // Attribute names.
		jsonKeys    []string           // JSON key names.
		columns     map[string]*Column // Map of columns.
	}
)

// Create new table.
func newTable(g *Gari, name string) *Table {
	return &Table{
		gari:        g,
		name:        name,
		columnNames: make([]string, 0),
		fieldNames:  make([]string, 0),
		jsonKeys:    make([]string, 0),
		columns:     make(map[string]*Column, 0),
	}
}

// Table name.
func (t *Table) Name() string {
	return t.name
}

// Build SELECT statement.
func (t *Table) Select(ptr any) *selectBuilder {
	return newSelectBuilder(t, ptr)
}

// Execute DDL SQL to migrate table.
func (t *Table) Migrate(ctx context.Context, db *sql.DB) error {
	// Build DDL SQL.
	stmt, err := newDdlBuilder(t).sql()
	fmt.Printf("MIGRATE(err=%v)\nSQL=%s\n", err, stmt)
	// Execute query.
	_, err = db.ExecContext(ctx, stmt)
	if err != nil {
		return err
	}
	return nil
}

/*
// sqlmock用の値をdriver.Valueのスライスとして取得する。
func (t *Table) DriverValues() []driver.Value {
	dbMap := t.toDbMap()
	values := make([]driver.Value, 0, len(dbMap))
	for _, column := range t.schema.columns {
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

// 列の追加
func (t *Table) addColumn(col *Column) {
	// 列名・属性名・JSONキー名の追加
	t.columnNames = append(t.columnNames, col.name)
	t.fieldNames = append(t.fieldNames, col.fieldName)
	// マップに追加する。
	t.columns[col.name] = col
	t.columns[col.fieldName] = col
	// テーブルへのポインタの追加
	col.table = t
}
