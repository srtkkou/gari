package gari

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

type (
	// Mock struct for testing.
	sampleStruct struct {
		Num  int
		Text string
	}
)

func TestSqlmock(t *testing.T) {
	// Define objects table.
	gari, err := New()
	require.NoError(t, err)
	sampleTable, err := gari.DefineTable("samples").
		AddInt64Column("num", func(col *Column) {
			col.SetDefaultInt64(0)
		}).
		AddStringColumn("text", func(col *Column) {
			col.SetSize(255)
			col.SetDefaultString("defaultString")
		}).
		Build()
	require.NoError(t, err)
	// Open sqlmock.
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	// Add expectation.
	rows := mock.NewRows([]string{"num", "text"}).
		AddRow(111, "str1").
		AddRow(222, "str2")
	mock.ExpectQuery(`SELECT .* FROM "samples" .*`).
		WillReturnRows(rows)
	// SELECT
	ctx := context.Background()
	samples := make([]sampleStruct, 0)
	err = sampleTable.Select(&samples).OrderAsc("num").All(ctx, db)
	require.NoError(t, err)
	// Check result.
	require.Equal(t, 2, len(samples))
	require.Equal(t, 111, samples[0].Num)
	require.Equal(t, "str1", samples[0].Text)
	require.Equal(t, 222, samples[1].Num)
	require.Equal(t, "str2", samples[1].Text)
}
