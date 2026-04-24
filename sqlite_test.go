package gari

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	// Define samples table.
	gari, err := New()
	require.NoError(t, err)
	table, err := gari.Table("samples").
		AddInt64Column("num",
			NotNull(), DefaultInt64(0)).
		AddStringColumn("text",
			NotNull(), Size(255), DefaultString("default")).
		Define()
	require.NoError(t, err)
	// Open SQLite
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys(1)")
	require.NoError(t, err)
	defer db.Close()
	// Migrate.
	ctx := context.Background()
	err = table.Migrate(ctx, db)
	require.NoError(t, err)
}
