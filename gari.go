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
	return &b
}
