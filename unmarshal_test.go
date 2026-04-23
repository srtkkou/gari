package gari

import (
	"database/sql"
	"testing"

	"github.com/izinin/json2msgpack"
	"github.com/stretchr/testify/require"
	msgpack "github.com/vmihailenco/msgpack/v5"
)

type (
	WantString struct {
		Value string
	}
	WantNullString struct {
		Value sql.NullString
	}
)

func TestUnmarshalMsgpack(t *testing.T) {
	str := "abc"
	tests := []struct {
		name    string
		input   string
		want    any
		isError bool
		output  any
	}{
		{
			name:    "OK:string",
			input:   `{"Value": "abc"}`,
			want:    WantString{Value: str},
			isError: false,
			output:  WantString{},
		},
		{
			name:  "OK:NullString",
			input: `{"Value": "abc"}`,
			want: WantNullString{
				Value: sql.NullString{Valid: true, String: str},
			},
			isError: false,
			output:  WantNullString{},
		},
		{
			name:  "OK:NullString:null",
			input: `{"Value": null}`,
			want: WantNullString{
				Value: sql.NullString{Valid: false},
			},
			isError: false,
			output:  WantNullString{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mp := json2msgpack.EncodeJSON([]byte(test.input))
			t.Logf("JSON=%s\n", test.input)
			t.Logf("Msgpack=%#x\n", mp)
			t.Logf("want=%v(%T)\n", test.want, test.want)
			t.Logf("output=%v(%T)\n", test.output, test.output)
			v := WantNullString{}
			t.Logf("v=%v(%T)\n", v, v)
			err := msgpack.Unmarshal(mp, &v)
			//err := msgpack.Unmarshal(mp, &(test.output))
			if test.isError {
				require.NotNil(t, err)
			} else {
				//t.Logf("got=%v(%T)\n", test.output, test.output)
				t.Logf("got=%v(%T)\n", v, v)
				require.NoError(t, err)
				//require.Equal(t, test.want, test.output)
				require.Equal(t, test.want, v)
			}
		})
	}
}
