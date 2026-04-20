package gari

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	g, err := New()
	require.NoError(t, err)
	tb := newTable(g, "users")
	idCol := newColumn("id")
	idCol.kind = kindString
	tb.columnNames = []string{idCol.name}
	tb.columns[idCol.name] = idCol
	require.NotEmpty(t, idCol.kind)
}
