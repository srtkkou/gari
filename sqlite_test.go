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
	type testModel struct {
		Model
		Num  int
		Text string
	}
	// Open SQLite
	db, err := sql.Open("sqlite", ":memory:?_foreign_keys(1)")
	require.NoError(t, err)
	defer db.Close()
	// Initialize gari.
	gari, err := Open(db)
	gari.Debug = tLog(t)
	gari.Info = tLog(t)
	gari.Warn = tLog(t)
	gari.Error = tLog(t)
	require.NoError(t, err)
	defer gari.Close()
	// Define table.
	table, err := gari.Table("test_models").
		AddInt64Column("num",
			NotNull(), DefaultInt64(0)).
		AddStringColumn("text",
			NotNull(), Size(255), DefaultString("DEFAULT")).
		Define()
	require.NoError(t, err)
	// Migrate.
	ctx := context.Background()
	err = table.Migrate(ctx)
	require.NoError(t, err)
	// INSERT.
	m1 := testModel{Num: 101, Text: "str1"}
	m2 := testModel{Num: 102, Text: "str2"}
	m3 := testModel{Num: 103, Text: "str3"}
	err = table.Insert().Values(&m1, &m2, &m3).Exec(ctx)
	require.NoError(t, err)
	// SELECT.
	models := make([]testModel, 0)
	err = table.Select().OrderAsc("num").Exec(ctx, func(r *Record) {
		m := testModel{}
		r.SetInt("id", &m.Id)
		r.SetTime("created_at", &m.CreatedAt)
		r.SetTime("updated_at", &m.UpdatedAt)
		r.SetNullTime("deleted_at", &m.DeletedAt)
		r.SetInt("num", &m.Num)
		r.SetString("text", &m.Text)
		models = append(models, m)
	})
	require.NoError(t, err)
	// Check result.
	require.Equal(t, 3, len(models))
	require.Equal(t, 1, models[0].Id)
	require.Equal(t, 101, models[0].Num)
	require.Equal(t, "str1", models[0].Text)
	require.Equal(t, 2, models[1].Id)
	require.Equal(t, 102, models[1].Num)
	require.Equal(t, "str2", models[1].Text)
	require.Equal(t, 3, models[2].Id)
	require.Equal(t, 103, models[2].Num)
	require.Equal(t, "str3", models[2].Text)
	// UPDATE.
	models[1].Num = 202
	models[1].Text = "STR2"
	models[2].Num = 203
	models[2].Text = "STR3"
	err = table.Update().Values(&models[1], &models[2]).Exec(ctx)
	require.NoError(t, err)
}
