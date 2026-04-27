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
		columns        map[string]*Column // Map of columns.
		insertBuilder  *insertBuilder     // INSERT builder.
		preparedInsert *sql.Stmt          // Prepared INSERT statement.
		updateBuilder  *updateBuilder     // UPDATE builder.
		preparedUpdate *sql.Stmt          // Prepared UPDATE statement.
	}
)

var (
	ErrInsert = errors.New("gari.ErrInsert")
	ErrUpdate = errors.New("gari.ErrUpdate")
)

// Create new table.
func newTable(g *Gari, name string) *Table {
	t := Table{
		gari:        g,
		name:        name,
		columnNames: make([]string, 0),
		fieldNames:  make([]string, 0),
		columns:     make(map[string]*Column, 0),
	}
	return &t
}

// Table name.
func (t *Table) Name() string {
	return t.name
}

// Build SELECT SQL statement.
func (t *Table) Select(
	ctx context.Context, db *sql.DB, ptr any,
) *selectBuilder {
	return newSelectBuilder(ctx, db, t, ptr)
}

// Execute INSERT SQL statement.
func (t *Table) Insert(
	ctx context.Context, db *sql.DB, ptrs ...any,
) (err error) {
	if len(ptrs) == 0 {
		return errs.Wrap(ErrInsert,
			errs.WithContext("SizeOfPtrs", len(ptrs)))
	}
	// Prepare insertBuilder.
	if t.insertBuilder == nil {
		t.insertBuilder = newInsertBuilder(t)
	}
	// Prepare INSERT SQL statement.
	query := t.insertBuilder.query
	if t.preparedInsert == nil {
		t.preparedInsert, err = db.PrepareContext(ctx, query)
		if err != nil {
			return errs.Wrap(ErrInsert, errs.WithCause(err),
				errs.WithContext("query", query))
		}
	}
	// Execute query.
	for _, ptr := range ptrs {
		// Build args.
		args, err := t.insertBuilder.buildArgs(ptr)
		if err != nil {
			return errs.Wrap(ErrInsert, errs.WithCause(err),
				errs.WithContext("query", query))
		}
		// Execute query.
		result, err := t.preparedInsert.ExecContext(ctx, args...)
		if err != nil {
			return errs.Wrap(ErrInsert, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
		count, err := result.RowsAffected()
		if err != nil {
			return errs.Wrap(ErrInsert, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args))
		}
		if count != 1 {
			return errs.Wrap(ErrInsert, errs.WithCause(err),
				errs.WithContext("query", query),
				errs.WithContext("args", args),
				errs.WithContext("rowsAffected", count))
		}
	}
	return nil
}

// Execute UPDATE SQL statement.
func (t *Table) Update(
	ctx context.Context, db *sql.DB, ptrs ...any,
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
func (t *Table) Migrate(ctx context.Context, db *sql.DB) error {
	// Build DDL SQL.
	ddl, err := newDdlBuilder(t).build()
	fmt.Printf("MIGRATE(err=%v)\nSQL=%s\n", err, ddl)
	// Execute query.
	_, err = db.ExecContext(ctx, ddl)
	if err != nil {
		return err
	}
	return nil
}

func (t *Table) Close() (err error) {
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
