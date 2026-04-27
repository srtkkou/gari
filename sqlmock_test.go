package gari

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestSqlmock(t *testing.T) {
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
	defer table.Close()
	// Open sqlmock.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	// Add DDL expectation.
	mock.ExpectExec(`CREATE TABLE IF NOT EXISTS samples(.*)`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	// Add INSERT expectations.
	preparedInsert := mock.ExpectPrepare(`INSERT INTO "samples"`)
	preparedInsert.ExpectExec().WithArgs(101, "str1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	preparedInsert.ExpectExec().WithArgs(102, "str2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	preparedInsert.ExpectExec().WithArgs(103, "str3").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// Add SELECT expectation.
	rows := mock.NewRows([]string{"num", "text"}).
		AddRow(101, "str1").
		AddRow(102, "str2").
		AddRow(103, "str3")
	mock.ExpectQuery(`SELECT .* FROM "samples" .*`).
		WillReturnRows(rows)
	// Migrate.
	ctx := context.Background()
	err = table.Migrate(ctx, db)
	require.NoError(t, err)
	// INSERT.
	s1 := sample{Num: 101, Text: "str1"}
	s2 := sample{Num: 102, Text: "str2"}
	s3 := sample{Num: 103, Text: "str3"}
	err = table.Insert(ctx, db, &s1, &s2, &s3)
	require.NoError(t, err)
	// SELECT.
	samples := make([]sample, 0)
	err = table.Select(ctx, db, &samples).OrderAsc("num").All()
	require.NoError(t, err)
	// Check result.
	require.Equal(t, 3, len(samples))
	require.Equal(t, 101, samples[0].Num)
	require.Equal(t, "str1", samples[0].Text)
	require.Equal(t, 102, samples[1].Num)
	require.Equal(t, "str2", samples[1].Text)
	require.Equal(t, 103, samples[2].Num)
	require.Equal(t, "str3", samples[2].Text)
}
