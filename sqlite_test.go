package gari_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/srtkkou/gari"
	"github.com/stretchr/testify/require"
)

func TestSqlite(t *testing.T) {
	// Define struct.
	type testModel struct {
		Id        int
		CreatedAt time.Time
		UpdatedAt time.Time
		DeletedAt sql.NullTime
		Num       int
		Text      string
	}
	// Open SQLite
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys(1)")
	require.NoError(t, err)
	defer db.Close()
	// Initialize gari.
	dialect := gari.SqliteDialect{}
	g, err := gari.Open(db, dialect)
	require.NoError(t, err)
	defer g.Close()
	// Define table.
	table, err := g.Table("test_models").
		IdColumn().
		TimeStampColumns().
		SoftDeleteColumn().
		Int64Column("num",
			gari.NotNull(), gari.DefaultInt64(0)).
		StringColumn("text",
			gari.NotNull(), gari.Length(255),
			gari.DefaultString("DEFAULT")).
		Define()
	require.NoError(t, err)
	// Migrate.
	ctx := context.Background()
	err = table.Migrate(ctx)
	require.NoError(t, err)
	// INSERT.
	models := make([]*testModel, 5)
	models[0] = &testModel{Num: 100, Text: "str0"}
	models[1] = &testModel{Num: 101, Text: "str1"}
	models[2] = &testModel{Num: 102, Text: "str2"}
	models[3] = &testModel{Num: 103, Text: "str3"}
	models[4] = &testModel{Num: 104, Text: "str4"}
	args := make([]any, len(models))
	for i := range args {
		args[i] = models[i]
	}
	err = table.Insert().Values(args...).Exec(ctx)
	require.NoError(t, err)
	// SELECT.
	results := make([]testModel, 0)
	//result := testModel{}
	err = g.Select(
		`SELECT "id", "created_at", "updated_at", "deleted_at",
		"num", "text" FROM "test_models" ORDER BY "id" ASC;`,
	).AssignTo(&results).Exec(ctx,
		gari.CamelToSnakeMapper(&results),
		/*
			func(r *gari.Record) {
				m := testModel{}
			r.SetInt("id", &m.Id)
			r.SetTime("created_at", &m.CreatedAt)
			r.SetTime("updated_at", &m.UpdatedAt)
			r.SetNullTime("deleted_at", &m.DeletedAt)
			r.SetInt("num", &m.Num)
			r.SetString("text", &m.Text)
			results = append(results, m)
			}
		*/
	)
	require.NoError(t, err)
	// Check result.
	require.Equal(t, len(models), len(results))
	for i := range len(results) {
		require.Equal(t, i+1, results[i].Id)
		require.Equal(t, i+100, results[i].Num)
		require.Equal(t, fmt.Sprintf("str%d", i), results[i].Text)
	}
	// UPDATE.
	results[1].Num = 202
	results[1].Text = "STR2"
	results[2].Num = 203
	results[2].Text = "STR3"
	err = table.Update().Values(&results[1], &results[2]).Exec(ctx)
	require.NoError(t, err)
	// DELETE.
	err = table.Delete().Values(&results[3]).Exec(ctx)
	require.NoError(t, err)
}
