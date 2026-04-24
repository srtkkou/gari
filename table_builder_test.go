package gari

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTableBuilder(t *testing.T) {
	g, _ := New()
	table, err := g.Table("users").
		AddStringColumn("xid", func(col *Column) {
			col.SetSize(32)
			col.SetDefaultString("")
		}).
		AddTimeColumn("created_at", func(col *Column) {
			col.AllowNull(true)
			col.SetDefaultNull()
		}).
		AddBoolColumn("ok", func(col *Column) {
			col.SetDefaultBool(false)
		}).
		Define()
	require.NoError(t, err)
	require.NotNil(t, table)
}
