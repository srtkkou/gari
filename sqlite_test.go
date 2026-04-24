package gari

import (
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	/*
		// Define objects table.
		gari, err := New()
		require.NoError(t, err)
		objectsTable, err := gari.DefineTable("objects").
			AddInt64Column("num", func(col *Column) {
				col.SetDefaultInt64(0)
			}).
			AddStringColumn("text", func(col *Column) {
				col.SetSize(255)
				col.SetDefaultString("defaultString")
			}).
			Build()
		require.NoError(t, err)
	*/
	// Open SQLite
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys(1)")
	require.NoError(t, err)
	defer db.Close()
}
