package gari

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableBuilder(t *testing.T) {
	g, _ := New()
	table, err := g.Table("users").
		AddStringColumn("xid",
			NotNull(), Size(32), DefaultString("")).
		AddTimeColumn("created_at",
			DefaultNull()).
		AddBoolColumn("ok",
			NotNull(), DefaultBool(false)).
		Define()
	require.NoError(t, err)
	require.NotNil(t, table)
}
