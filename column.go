package gari

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/goark/errs"
)

type (
	// 列
	Column struct {
		table        *Table            // テーブル
		name         string            // テーブル列名
		attrName     string            // 属性名
		jsonKey      string            // JSONキー名
		kind         kind              // Type
		size         int               // 長さ
		allowNull    bool              // NULL型の許容
		value        []byte            // 現在の値
		defaultValue []byte            // 既定値
		rules        []validation.Rule // 検証ルール
		err          error             // エラー
		//		dbKind       sql.ColumnType    // DB上の型
	}
	// Option of column.
	ColumnOption func(*Column)
)

var (
	// 初期値のNULLエラー
	ErrDefaultValueNull = errors.New("orm.ErrDefaultValueNull")
	// 初期値の型エラー
	ErrDefaultValueKind = errors.New("orm.ErrDefaultValueKind")
	// 初期値の長さエラー
	ErrDefaultValueSize = errors.New("orm.ErrDefaultValueSize")
)

// 列の新規作成
func newColumn(name string) *Column {
	return &Column{
		name:         name,
		attrName:     snakeToUpperCamelCase(name),
		jsonKey:      snakeToLowerCamelCase(name),
		kind:         "",
		size:         -1,
		allowNull:    true,
		value:        []byte{},
		defaultValue: []byte{},
		rules:        make([]validation.Rule, 0),
	}
}

// 文字列化
func (c *Column) String() string {
	var b strings.Builder
	b.WriteString(`{"column":"`)
	b.WriteString(c.name)
	b.WriteString(`","attrName":"`)
	b.WriteString(c.attrName)
	b.WriteString(`","jsonKey":"`)
	b.WriteString(c.jsonKey)
	b.WriteString(`","kind":"`)
	b.WriteString(c.kind)
	b.WriteString(`","size":`)
	b.WriteString(strconv.Itoa(c.size))
	b.WriteString(`,"allowNull":`)
	b.WriteString(strconv.FormatBool(c.allowNull))
	b.WriteString(`}`)
	return b.String()
}

// Set field name.
func FieldName(name string) ColumnOption {
	return func(c *Column) {
		c.attrName = name
	}
}

// JSONキー名
func (c *Column) SetJsonKey(key string) {
	c.jsonKey = key
}

// Set size of column.
func Size(size int) ColumnOption {
	return func(c *Column) {
		c.size = size
	}
}

// Disallow NULL for column.
func NotNull() ColumnOption {
	return func(c *Column) {
		c.allowNull = false
	}
}

// Set default value NULL.
func DefaultNull() ColumnOption {
	return func(c *Column) {
		if !c.allowNull {
			err := errs.Wrap(ErrDefaultValueNull,
				errs.WithContext("columnName", c.name),
				errs.WithContext("allowNull", c.allowNull))
			c.addError(err)
			return
		}
		c.value = msgpackNil()
	}
}

// Set default bool value.
func DefaultBool(b bool) ColumnOption {
	return func(c *Column) {
		err := c.validateKind(kindBool)
		if err != nil {
			c.addError(err)
			return
		}
		blob, err := encodeNullBool(b)
		if err != nil {
			c.addError(err)
			return
		}
		c.value = blob
	}
}

// Set default string value.
func DefaultString(str string) ColumnOption {
	return func(c *Column) {
		err := c.validateKind(kindString)
		if err != nil {
			c.addError(err)
			return
		}
		if c.size < len(str) {
			err = errs.Wrap(ErrDefaultValueSize,
				errs.WithContext("input", str),
				errs.WithContext("size", c.size))
			c.addError(err)
			return
		}
		blob, err := encodeNullString(str)
		if err != nil {
			c.addError(err)
			return
		}
		c.value = blob
	}
}

// Set default int value.
func DefaultInt64(num int64) ColumnOption {
	return func(c *Column) {
		err := c.validateKind(kindInt64)
		if err != nil {
			c.addError(err)
			return
		}
		blob, err := encodeNullInt64(num)
		if err != nil {
			c.addError(err)
			return
		}
		c.value = blob
	}
}

// 検証ルール群の追加
func (c *Column) AddRules(rules ...validation.Rule) {
	c.rules = append(c.rules, rules...)
}

// 型のチェック
func (c *Column) validateKind(kinds ...string) error {
	if !slices.Contains(kinds, c.kind) {
		return errs.Wrap(ErrDefaultValueKind,
			errs.WithContext("columnName", c.name),
			errs.WithContext("kind", c.kind))
	}
	return nil
}

// エラーの追加
func (c *Column) addError(err error) {
	c.err = errors.Join(c.err, err)
}
