package gari

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSqlmock(t *testing.T) {
	// Define struct.
	type testModel struct {
		Model
		Num  int
		Text string
	}
	// Open sqlmock.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	// Define test_models table.
	gari, err := Open(db)
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
	defer table.Close()
	// Add DDL expectation.
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS test_models(.*)`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Add INSERT expectations.
	preparedInsert := mock.ExpectPrepare(`INSERT INTO "test_models"`)
	preparedInsert.ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))
	preparedInsert.ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))
	preparedInsert.ExpectExec().
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Add SELECT expectation.
	columns := []string{
		"id", "created_at", "updated_at", "deleted_at",
		"num", "text",
	}
	rows := mock.NewRows(columns).
		AddRow(1, time20111213(), time20111213(), nullTime(), 101, "str1").
		AddRow(2, time20111213(), time20111213(), nullTime(), 102, "str2").
		AddRow(3, time20111213(), time20111213(), nullTime(), 103, "str3")
	mock.ExpectQuery(`SELECT .* FROM "test_models" .*`).
		WillReturnRows(rows)
	// Migrate.
	ctx := context.Background()
	err = table.Migrate(ctx)
	require.NoError(t, err)
	// INSERT.
	m1 := testModel{Num: 101, Text: "str1"}
	m2 := testModel{Num: 102, Text: "str2"}
	m3 := testModel{Num: 103, Text: "str3"}
	err = table.Insert(ctx, &m1, &m2, &m3)
	require.NoError(t, err)
	// SELECT.
	models := make([]testModel, 0)
	err = table.Select(ctx, &models).OrderAsc("num").All()
	require.NoError(t, err)
	// Check result.
	require.Equal(t, 3, len(models))
	require.Equal(t, 101, models[0].Num)
	require.Equal(t, "str1", models[0].Text)
	require.Equal(t, 102, models[1].Num)
	require.Equal(t, "str2", models[1].Text)
	require.Equal(t, 103, models[2].Num)
	require.Equal(t, "str3", models[2].Text)
}
