package gari

import (
	"testing"

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
		gari, err := New()
		require.NoError(t, err)
		t.Run(test.name, func(t *testing.T) {
			table, err := gari.Table("test").
				AddBoolColumn("is_ok",
					DefaultBool(test.input)).
				Define()
			require.NoError(t, err)
			col := table.columns["is_ok"]
			require.NotEmpty(t, col.defaultValue)
		})
	}
}
