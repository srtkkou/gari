package gari

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestColumnDefaultBool(t *testing.T) {
	tests := []struct {
		name  string
		input bool
		//want any
	}{
		{
			name:  "OK:true",
			input: true,
		},
		{
			name:  "OK:false",
			input: false,
		},
	}
	for _, test := range tests {
		// Open sqlmock.
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		// Use gari.
		g, err := Open(db)
		require.NoError(t, err)
		defer g.Close()
		t.Run(test.name, func(t *testing.T) {
			table, err := g.Table("test").
				Int64Column("id", PrimaryKey(true), NotNull()).
				BoolColumn("is_ok", DefaultBool(test.input)).
				Define()
			require.NoError(t, err)
			col := table.column("is_ok")
			require.NotNil(t, col)
			require.NotEmpty(t, col.defaultValue)
		})
	}
}
