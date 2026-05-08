package gari

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestTableBuilder(t *testing.T) {
	// Open sqlmock.
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	// Use gari.
	g, err := Open(db)
	require.NoError(t, err)
	defer g.Close()
	// Define table.
	table, err := g.Table("users").
		Int64Column("id", PrimaryKey(true), NotNull()).
		StringColumn("xid",
			NotNull(), Length(32), DefaultString("")).
		TimeColumn("created_at",
			DefaultNull()).
		Int64Column("num",
			NotNull(), DefaultInt64(123)).
		Define()
	require.NoError(t, err)
	require.NotNil(t, table)
}
