package gari

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/glebarez/go-sqlite"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	// Define struct.
	type sample struct {
		Num  int
		Text string
	}
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
	// INSERT.
	s1 := sample{Num: 101, Text: "str1"}
	s2 := sample{Num: 102, Text: "str2"}
	err = table.Insert(ctx, db, &s1, &s2)
	require.NoError(t, err)
	// SELECT.
	samples := make([]sample, 0)
	err = table.Select(ctx, db, &samples).OrderAsc("num").All()
	require.NoError(t, err)
	// Check result.
	require.Equal(t, 2, len(samples))
	require.Equal(t, 101, samples[0].Num)
	require.Equal(t, "str1", samples[0].Text)
	require.Equal(t, 102, samples[1].Num)
	require.Equal(t, "str2", samples[1].Text)
}
