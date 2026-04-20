package gari

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSnakeToLowerCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OK:snake_case_sample",
			input: "snake_case_sample",
			want:  "snakeCaseSample",
		},
		{
			name:  "OK:xid",
			input: "xid",
			want:  "xid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := snakeToLowerCamelCase(test.input)
			require.Equal(t, test.want, got)
		})
	}
}

func TestSnakeToUpperCamelCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "OK:snake_case_sample",
			input: "snake_case_sample",
			want:  "SnakeCaseSample",
		},
		{
			name:  "OK:xid",
			input: "xid",
			want:  "Xid",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := snakeToUpperCamelCase(test.input)
			require.Equal(t, test.want, got)
		})
	}
}
