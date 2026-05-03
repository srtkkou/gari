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
