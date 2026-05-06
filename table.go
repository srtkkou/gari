package gari

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/goark/errs"
)

type (
	// Table
	Table struct {
		gari           *Gari              // Pointer to gari config.
		name           string             // Table name.
		columnNames    []string           // Column names.
		fieldNames     []string           // Attribute names.
		columns        map[string]*column // Map of columns.
		insertBuilder  *insertBuilder     // INSERT builder.
		updateBuilder  *updateBuilder     // UPDATE builder.
		preparedUpdate *sql.Stmt          // Prepared UPDATE statement.
	}
)

var (
	ErrUpdate = errors.New("gari.ErrUpdate")
)

// Create new table.
func newTable(g *Gari, name string) *Table {
	t := Table{
		gari:        g,
		name:        name,
		columnNames: make([]string, 0),
		fieldNames:  make([]string, 0),
		columns:     make(map[string]*column, 0),
	}
	return &t
}

// Table name.
func (t *Table) Name() string {
	return t.name
}

// Build SELECT SQL statement.
func (t *Table) Select(
	ctx context.Context, ptr any,
) *selectBuilder {
	return newSelectBuilder(ctx, t, ptr)
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
func (t *Table) Update(
	ctx context.Context, ptrs ...any,
) (err error) {
	if len(ptrs) == 0 {
		return errs.Wrap(ErrUpdate,
			errs.WithContext("sizeOfPtrs", len(ptrs)))
	}
	// Prepare updateBuilder.
	if t.updateBuilder == nil {
		t.updateBuilder = newUpdateBuilder(t)
	}
	// Prepare UPDATE SQL statement.
	db := t.gari.db
	query := t.updateBuilder.query
	if t.preparedUpdate == nil {
		t.preparedUpdate, err = db.PrepareContext(ctx, query)
		if err != nil {
			return errs.Wrap(ErrUpdate, errs.WithCause(err))
		}
	}
	// Execute query for each struct.
	for _, ptr := range ptrs {
		// Get args,
		args, err := t.updateBuilder.buildArgs(ptr)
		if err != nil {
			return errs.Wrap(ErrUpdate, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
		// Execute query.
		result, err := t.preparedUpdate.ExecContext(ctx, args...)
		if err != nil {
			return errs.Wrap(ErrUpdate, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
		count, err := result.RowsAffected()
		if err != nil {
			return errs.Wrap(ErrUpdate, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
		if count != 1 {
			return errs.Wrap(ErrUpdate,
				errs.WithContext("rowsAffected", count),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
	}
	return nil
}

// Execute DDL SQL to migrate table.
func (t *Table) Migrate(ctx context.Context) error {
	// Build DDL SQL.
	ddl, err := newDdlBuilder(t).build()
	fmt.Printf("MIGRATE(err=%v)\nSQL=%s\n", err, ddl)
	// Execute query.
	db := t.gari.db
	_, err = db.ExecContext(ctx, ddl)
	if err != nil {
		return err
	}
	return nil
}

// Close prepared statements.
func (t *Table) closeTable() (err error) {
	if t.insertBuilder != nil {
		t.insertBuilder.closePreparedStmt()
	}
	if t.preparedUpdate != nil {
		if err = t.preparedUpdate.Close(); err != nil {
			return err
		}
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

// Add column to table
func (t *Table) addColumn(col *column) {
	// Add column name and field name to slice.
	t.columnNames = append(t.columnNames, col.name)
	t.fieldNames = append(t.fieldNames, col.fieldName)
	// Add column pointer to map.
	t.columns[col.name] = col
	t.columns[col.fieldName] = col
	// Add pointer to table on column.
	col.table = t
}
