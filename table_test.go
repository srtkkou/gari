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
	g, err := Open(db)
	require.NoError(t, err)
	defer g.Close()
	// Define table.
	tb := newTable(g, "users")
	idCol := newColumn("id")
	idCol.kind = kindString
	tb.columns = []*column{idCol}
	tb.columnMap[idCol.name] = idCol
	require.NotEmpty(t, idCol.kind)
}
