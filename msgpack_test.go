package gari

import (
	"database/sql"
	"testing"
	"time"

	j2m "github.com/izinin/json2msgpack"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeBool(t *testing.T) {
	tests := []struct {
		name          string
		input         bool
		wantEncodeErr bool
		wantDecodeErr bool
		want          sql.NullBool
	}{
		{
			name:          "OK:true",
			input:         true,
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullBool{Valid: true, Bool: true},
		},
		{
			name:          "OK:false",
			input:         false,
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullBool{Valid: true, Bool: false},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeBool(test.input)
			if test.wantEncodeErr {
				require.Error(t, err1)
			} else {
				require.NoError(t, err1)
				require.NotEmpty(t, blob)
			}
			// Decode.
			nb, err2 := decodeNullBool(blob)
			if test.wantDecodeErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
			}
			require.Equal(t, test.want, nb)
		})
	}
}

func TestDecodeNullBool(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		want    sql.NullBool
	}{
		{
			name:    "OK:true",
			input:   "true",
			wantErr: false,
			want:    sql.NullBool{Valid: true, Bool: true},
		},
		{
			name:    "OK:false",
			input:   "false",
			wantErr: false,
			want:    sql.NullBool{Valid: true, Bool: false},
		},
		{
			name:    "OK:nil",
			input:   "null",
			wantErr: false,
			want:    sql.NullBool{Valid: false, Bool: false},
		},
		{
			name:    "NG:string",
			input:   `"ABC"`,
			wantErr: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			blob := j2m.EncodeJSON([]byte(test.input))
			t.Logf("JSON=%s, msgpack=%#x\n", test.input, blob)
			// Decode.
			nb, err := decodeNullBool(blob)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.want, nb)
			}
		})
	}
}

func TestEncodeDecodeString(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantEncodeErr bool
		wantDecodeErr bool
		want          sql.NullString
	}{
		{
			name:          "OK:string",
			input:         "abcd",
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullString{Valid: true, String: "abcd"},
		},
		{
			name:          "OK:empty string",
			input:         "",
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullString{Valid: true, String: ""},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeString(test.input)
			if test.wantEncodeErr {
				require.Error(t, err1)
			} else {
				require.NoError(t, err1)
				require.NotEmpty(t, blob)
			}
			// Decode.
			ns, err2 := decodeNullString(blob)
			if test.wantDecodeErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				require.Equal(t, test.want, ns)
			}
		})
	}
}

func TestEncodeDecodeTime(t *testing.T) {
	tests := []struct {
		name          string
		input         time.Time
		wantEncodeErr bool
		wantDecodeErr bool
		want          sql.NullTime
	}{
		{
			name:          "OK:time.Time",
			input:         time20111213(),
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullTime{Valid: true, Time: time20111213()},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeTime(test.input)
			if test.wantEncodeErr {
				require.Error(t, err1)
			} else {
				require.NoError(t, err1)
				require.NotEmpty(t, blob)
			}
			// Decode.
			nt, err2 := decodeNullTime(blob)
			if test.wantDecodeErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				require.Equal(t, test.want, nt)
			}
		})
	}
}

func TestEncodeDecodeInt64(t *testing.T) {
	tests := []struct {
		name          string
		input         int64
		wantEncodeErr bool
		wantDecodeErr bool
		want          sql.NullInt64
	}{
		{
			name:          "OK:int64",
			input:         int64(12),
			wantEncodeErr: false,
			wantDecodeErr: false,
			want:          sql.NullInt64{Valid: true, Int64: 12},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeInt64(test.input)
			if test.wantEncodeErr {
				require.Error(t, err1)
			} else {
				require.NoError(t, err1)
				require.NotEmpty(t, blob)
			}
			// Decode.
			ni, err2 := decodeNullInt64(blob)
			if test.wantDecodeErr {
				require.Error(t, err2)
			} else {
				require.NoError(t, err2)
				require.Equal(t, test.want, ni)
			}
		})
	}
}
