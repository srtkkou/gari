package gari

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestColumnDefaultString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "OK:string",
			input: "abcd",
		},
		{
			name:  "OK:empty string",
			input: "",
		},
	}
	for _, test := range tests {
		// Open sqlmock.
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		// Use gari.
		dialect := SqliteDialect{}
		g, err := Open(db, dialect)
		require.NoError(t, err)
		defer g.Close()
		t.Run(test.name, func(t *testing.T) {
			table, err := g.Table("test").
				Int64Column("id", PrimaryKey(true), NotNull()).
				StringColumn("text", DefaultString(test.input)).
				Define()
			require.NoError(t, err)
			col := table.columnByName("text")
			require.NotNil(t, col)
			require.Equal(t, test.input, col.defaultValue)
		})
	}
}
