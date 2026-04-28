package gari

import (
	"errors"

	"github.com/goark/errs"
)

type (
	// 設定
	Gari struct {
		stringQuote string // 文字列のクォーテーション
		columnQuote string // 列名のクォーテーション
	}
	// オプション関数
	OptionFunc func(*Gari) error
)

var (
	// オプションエラー
	ErrOption = errors.New("gari.ErrOption")
)

// Initialize.
func New(fns ...OptionFunc) (*Gari, error) {
	g := Gari{
		stringQuote: `'`,
		columnQuote: `"`,
	}
	for _, fn := range fns {
		if err := fn(&g); err != nil {
			err = errs.Wrap(ErrOption, errs.WithCause(err))
			return nil, err
		}
	}
	return &g, nil
}

// Start table builder sequence.
func (g *Gari) Table(name string) *TableBuilder {
	b := TableBuilder{
		table: newTable(g, name),
	}
	// Add id column.
	b.AddInt64Column("id", func(c *Column) {
		c.primary = true
		c.autoIncrement = true
		c.notNull = true
	})
	// Add timestamp columns.
	b.AddTimeColumn("created_at", NotNull())
	b.AddTimeColumn("updated_at", NotNull())
	b.AddTimeColumn("deleted_at", DefaultNull())
	return &b
}
