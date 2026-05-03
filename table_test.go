package gari

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestTable(t *testing.T) {
	// Open sqlmock.
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	// Use gari.
	g, err := New(db)
	require.NoError(t, err)
	defer g.Close()
	// Define table.
	tb := newTable(g, "users")
	idCol := newColumn("id")
	idCol.kind = kindString
	tb.columnNames = []string{idCol.name}
	tb.columns[idCol.name] = idCol
	require.NotEmpty(t, idCol.kind)
}
