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
	sampleTable := gari.Table("samples").
		AddInt64Column("num", func(col *Column) {
			col.SetDefaultInt64(0)
		}).
		AddStringColumn("text", func(col *Column) {
			col.SetSize(255)
			col.SetDefaultString("defaultString")
		}).
		MustDefine()
	// Open SQLite
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys(1)")
	require.NoError(t, err)
	defer db.Close()
	// Migrate.
	ctx := context.Background()
	err = sampleTable.Migrate(ctx, db)
	require.NoError(t, err)
}
