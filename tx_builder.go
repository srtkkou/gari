package gari

import ()

type (
	// Transaction builder.
	txBuilder struct {
		gari *Gari
	}
)

func newTxBuilder(g *Gari) *txBuilder {
	b := &txBuilder{
		gari: g,
	}
	return b
}

func (b *txBuilder) Commit() error {
	return nil
}

func (b *txBuilder) Rollback() error {
	return nil
}
