package gari

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetGetBool(t *testing.T) {
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
		col := newColumn("bool_column")
		col.kind = kindBool
		t.Run(test.name, func(t *testing.T) {
			col.SetDefaultBool(test.input)
			require.NotEmpty(t, col.value)
		})
	}
}
