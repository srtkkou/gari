package gari

import (
	"strings"
)

type (
	// テーブル
	Table struct {
		gari        *Gari              // 設定への参照
		name        string             // 名前
		columnNames []string           // 列名群
		attrNames   []string           // 属性名群
		jsonKeys    []string           // JSONキー名群
		columns     map[string]*Column // 列情報マップ
	}
)

/*
var (
	// INSERT文作成エラー
	ErrTableInsertQuery = errors.New("orm.ErrTableInsertQuery")
	// UPDATE文作成エラー
	ErrTableUpdateQuery = errors.New("orm.ErrTableUpdateQuery")

	// SQLビルダー
	sqlBuilder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)
*/

// 空テーブルの作成
func newTable(g *Gari, name string) *Table {
	return &Table{
		gari:        g,
		name:        name,
		columnNames: make([]string, 0),
		attrNames:   make([]string, 0),
		jsonKeys:    make([]string, 0),
		columns:     make(map[string]*Column, 0),
	}
}

// テーブル名
func (t *Table) Name() string {
	return t.name
}

func (t *Table) SelectSql() string {
	var b strings.Builder
	b.WriteString("SELECT ")
	for i, name := range t.columnNames {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(t.gari.columnQuote)
		b.WriteString(t.name)
		b.WriteString(t.gari.columnQuote)
		b.WriteString(".")
		b.WriteString(t.gari.columnQuote)
		b.WriteString(name)
		b.WriteString(t.gari.columnQuote)
	}
	b.WriteString(" FROM ")
	b.WriteString(t.gari.columnQuote)
	b.WriteString(t.name)
	b.WriteString(t.gari.columnQuote)
	return b.String()
}

func (t *Table) Cells() []any {
	cells := make([]any, len(t.columnNames))
	for i, name := range t.columnNames {
		col := t.columns[name]
		cell := newCell(col)
		cells[i] = cell
	}
	return cells
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
	t.attrNames = append(t.attrNames, col.attrName)
	t.jsonKeys = append(t.jsonKeys, col.jsonKey)
	// マップに追加する。
	t.columns[col.name] = col
	t.columns[col.attrName] = col
	t.columns[col.jsonKey] = col
	// テーブルへのポインタの追加
	col.table = t
}
