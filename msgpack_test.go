package gari

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeBool(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		encodeErr error
		decodeErr error
		isNil     bool
	}{
		{
			name:      "OK:true",
			input:     true,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:false",
			input:     false,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:nil",
			input:     nil,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
		{
			name:      "OK:sql.NullBool:true",
			input:     sql.NullBool{Valid: true, Bool: true},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:sql.NullBool:false",
			input:     sql.NullBool{Valid: true, Bool: false},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:sql.NullBool:NULL",
			input:     sql.NullBool{Valid: false},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeNullBool(test.input)
			require.Equal(t, test.encodeErr, err1)
			if err1 == nil {
				require.NotEmpty(t, blob)
			}
			// Decode.
			nb, err2 := decodeNullBool(blob)
			require.Equal(t, test.decodeErr, err2)
			require.Equal(t, test.isNil, !nb.Valid)
			if nb.Valid {
				switch tv := test.input.(type) {
				case bool:
					require.Equal(t, tv, nb.Bool)
				case sql.NullBool:
					require.Equal(t, tv, nb)
				}
			}
		})
	}
}

func TestEncodeDecodeString(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		encodeErr error
		decodeErr error
		isNil     bool
	}{
		{
			name:      "OK:string",
			input:     "abcd",
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:nil",
			input:     nil,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
		{
			name:      "OK:NullString",
			input:     sql.NullString{Valid: true, String: "efgh"},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:NullString:NULL",
			input:     sql.NullString{Valid: false},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeNullString(test.input)
			require.Equal(t, test.encodeErr, err1)
			if err1 == nil {
				require.NotEmpty(t, blob)
			}
			// Decode.
			ns, err2 := decodeNullString(blob)
			require.Equal(t, test.decodeErr, err2)
			require.Equal(t, test.isNil, !ns.Valid)
			if ns.Valid {
				switch tv := test.input.(type) {
				case string:
					require.Equal(t, tv, ns.String)
				case sql.NullString:
					require.Equal(t, tv, ns)
				}
			}
		})
	}
}

func TestEncodeDecodeTime(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		encodeErr error
		decodeErr error
		isNil     bool
	}{
		{
			name:      "OK:time.Time",
			input:     time20111213(),
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:nil",
			input:     nil,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
		{
			name:      "OK:sql.NullTime",
			input:     sql.NullTime{Time: time20111213(), Valid: true},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:sql.NullTime:NULL",
			input:     sql.NullTime{Time: time20111213(), Valid: false},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeNullTime(test.input)
			require.Equal(t, test.encodeErr, err1)
			if err1 == nil {
				require.NotEmpty(t, blob)
			}
			// Decode.
			nt, err2 := decodeNullTime(blob)
			require.Equal(t, test.decodeErr, err2)
			require.Equal(t, test.isNil, !nt.Valid)
			if nt.Valid {
				switch tv := test.input.(type) {
				case time.Time:
					require.Equal(t, tv, nt.Time)
				case sql.NullTime:
					require.Equal(t, tv, nt)
				}
			}
		})
	}
}

func TestEncodeDecodeInt64(t *testing.T) {
	tests := []struct {
		name      string
		input     any
		encodeErr error
		decodeErr error
		isNil     bool
	}{
		{
			name:      "OK:int",
			input:     11,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:int16",
			input:     int16(12),
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:int32",
			input:     int32(13),
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:int64",
			input:     int64(14),
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:nil",
			input:     nil,
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
		{
			name:      "OK:sql.NullInt16",
			input:     sql.NullInt16{Int16: int16(15), Valid: true},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     false,
		},
		{
			name:      "OK:sql.NullInt16:NULL",
			input:     sql.NullInt16{Valid: false},
			encodeErr: nil,
			decodeErr: nil,
			isNil:     true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Encode.
			blob, err1 := encodeNullInt64(test.input)
			require.Equal(t, test.encodeErr, err1)
			if err1 == nil {
				require.NotEmpty(t, blob)
			}
			// Decode.
			ni, err2 := decodeNullInt64(blob)
			require.Equal(t, test.decodeErr, err2)
			require.Equal(t, test.isNil, !ni.Valid)
			if ni.Valid {
				switch tv := test.input.(type) {
				case int:
					require.Equal(t, int64(tv), ni.Int64)
				case int16:
					require.Equal(t, int64(tv), ni.Int64)
				case int32:
					require.Equal(t, int64(tv), ni.Int64)
				case int64:
					require.Equal(t, tv, ni.Int64)
				case sql.NullInt16:
					require.Equal(t, int64(tv.Int16), ni.Int64)
				case sql.NullInt32:
					require.Equal(t, int64(tv.Int32), ni.Int64)
				case sql.NullInt64:
					require.Equal(t, tv.Int64, ni.Int64)
				}
			}
		})
	}
}
