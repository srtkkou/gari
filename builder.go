package gari

import (
	"errors"
)

type (
	// テーブル定義ビルダー
	Builder struct {
		table *Table // テーブル
	}
)

// Bool型の列の追加
func (b *Builder) AddBoolColumn(
	name string, fn func(col *Column),
) *Builder {
	return b.addColumnWithKind(name, kindBool, fn)
}

// String型の列の追加
func (b *Builder) AddStringColumn(
	name string, fn func(col *Column),
) *Builder {
	return b.addColumnWithKind(name, kindString, fn)
}

// Time型の列の追加
func (b *Builder) AddTimeColumn(
	name string, fn func(col *Column),
) *Builder {
	return b.addColumnWithKind(name, kindTime, fn)
}

// 型を指定して列を追加する。
func (b *Builder) addColumnWithKind(
	name string, k kind, fn func(col *Column),
) *Builder {
	col := newColumn(name)
	col.kind = k
	fn(col)
	b.table.addColumn(col)
	return b
}

/*
// 整数型の列の追加
func (b *Builder) AddIntColumn(
	name string, num int64, rules ...validation.Rule,
) *Builder {
	b.columns = append(b.columns, name)
	b.kinds[name] = kindInt
	b.values[name] = kindInt.encodeInt64(num)
	// 検証ルールの追加
	b.validations[name] = rules
	return b
}
*/

// テーブルを作成する。
func (b *Builder) Build() (*Table, error) {
	// エラーの検証
	var err error
	for _, name := range b.table.columnNames {
		col := b.table.columns[name]
		if col.err != nil {
			err = errors.Join(err, col.err)
		}
	}
	if err != nil {
		return nil, err
	}
	return b.table, nil
}

/*
// テーブル列の完全修飾名の一覧を得る。
func (s *) FullColumnNames() []string {
	names := make([]string, len(s.columns))
	for i, column := range s.columns {
		names[i] = fmt.Sprintf("%s.%s", s.name, column)
	}
	return names
}

// 列の型がBoolか？
func (s *) IsBool(name string) bool {
	if k, ok := s.kinds[name]; ok {
		return k == kindBool
	}
	return false
}

// 列の型がInt64か？
func (s *) IsInt64(name string) bool {
	if k, ok := s.kinds[name]; ok {
		return k == kindInt
	}
	return false
}

// 列の型がFloatか？
func (s *) IsFloat(name string) bool {
	if k, ok := s.kinds[name]; ok {
		return k == kindFloat
	}
	return false
}

// 列の型がStringか？
func (s *) IsString(name string) bool {
	if k, ok := s.kinds[name]; ok {
		return k == kindString
	}
	return false
}

// 列の型がlib.Timeか？
func (s *) IsTime(name string) bool {
	if k, ok := s.kinds[name]; ok {
		return k == kindTime
	}
	return false
}
*/
