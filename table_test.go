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
	dialect := SqliteDialect{}
	g, err := Open(db, dialect)
	require.NoError(t, err)
	defer g.Close()
	// Define table.
	tb := newTable(g, "users")
	idCol := newColumn(tb, "id")
	idCol.kind = kindString
	tb.columns = []*column{idCol}
	tb.columnMap[idCol.Name] = idCol
	require.NotEmpty(t, idCol.kind)
}
