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

// 初期化
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

// テーブル定義の開始宣言
func (g *Gari) DefineTable(name string) *Builder {
	b := Builder{
		table: newTable(g, name),
	}
	return &b
}
