package gari

import (
	"bytes"
	"database/sql"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

// Msgpack encoded value bytes.
type encoded []byte

var (
	ErrMsgpackEncode    = errors.New("gari.msgpack.ErrMsgpackEncode")
	ErrMsgpackDecode    = errors.New("gari.ErrMsgpackDecode")
	ErrMsgpackInputType = errors.New("gari.ErrMsgpackInputType")
	ErrMsgpackArraySize = errors.New("gari.ErrMsgpackArraySize")
	ErrMsgpackMapSize   = errors.New("gari.ErrMsgpackMapSize")
	ErrMsgpackMapToType = errors.New("gari.ErrMsgpackMapToType")
)

// Encode bool to msgpack.
func encodeBool(b bool) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err = enc.EncodeBool(b)
	if err != nil {
		err = errs.Wrap(ErrMsgpackEncode,
			errs.WithCause(err),
			errs.WithContext("input", b))
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode string to msgpack.
func encodeString(str string) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err = enc.EncodeString(str)
	if err != nil {
		err = errs.Wrap(ErrMsgpackEncode,
			errs.WithCause(err),
			errs.WithContext("input", str))
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode time.Time to msgpack.
func encodeTime(tm time.Time) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err = enc.EncodeTime(tm)
	if err != nil {
		err = errs.Wrap(ErrMsgpackEncode,
			errs.WithCause(err),
			errs.WithContext("input", tm))
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode int64 to msgpack.
func encodeInt64(num int64) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	err = enc.EncodeInt64(num)
	if err != nil {
		err = errs.Wrap(ErrMsgpackEncode,
			errs.WithCause(err),
			errs.WithContext("input", num))
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode to sql.NullBool value.
func decodeNullBool(blob []byte) (sql.NullBool, error) {
	nb := sql.NullBool{Valid: false}
	// Decode NULL.
	if slices.Equal(blob, msgpackNil()) {
		return nb, nil
	}
	// Decode bool.
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	b, err := dec.DecodeBool()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return nb, err
	}
	nb.Valid = true
	nb.Bool = b
	return nb, nil
}

// Decode to sql.NullString value.
func decodeNullString(blob []byte) (sql.NullString, error) {
	ns := sql.NullString{Valid: false}
	// Decode NULL.
	if slices.Equal(blob, msgpackNil()) {
		return ns, nil
	}
	// Decode string.
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	str, err := dec.DecodeString()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return ns, err
	}
	ns.Valid = true
	ns.String = str
	return ns, nil
}

// Decode to sql.NullTime value.
func decodeNullTime(blob []byte) (sql.NullTime, error) {
	nt := sql.NullTime{Valid: false}
	// Decode NULL.
	if slices.Equal(blob, msgpackNil()) {
		return nt, nil
	}
	// Decode time.Time.
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	t, err := dec.DecodeTime()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return nt, err
	}
	nt.Valid = true
	nt.Time = t.UTC()
	return nt, nil
}

// Decode to sql.NullInt64 value.
func decodeNullInt64(blob []byte) (sql.NullInt64, error) {
	ni := sql.NullInt64{Valid: false}
	// Decode NULL.
	if slices.Equal(blob, msgpackNil()) {
		return ni, nil
	}
	// Decode int64.
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	num, err := dec.DecodeInt64()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return ni, err
	}
	ni.Valid = true
	ni.Int64 = num
	return ni, nil
}

/*
func splitToMap(blob []byte) (map[string]encoded, error) {
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	// Get map length.
	size, err := dec.DecodeMapLen()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return nil, err
	}
	// Store raw bytes into map.
	m := make(map[string]encoded, size)
	for range size {
		key, err := dec.DecodeString()
		if err != nil {
			err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
				errs.WithContext("input", blob))
			return nil, err
		}
		blob, err := dec.DecodeRaw()
		if err != nil {
			err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
				errs.WithContext("input", blob))
			return nil, err
		}
		m[key] = encoded(blob)
	}
	return m, nil
}
*/

// NULL in msgpack format.
func msgpackNil() []byte {
	return []byte{0xc0}
}

// Stringify encoded bytes.
func (e encoded) String() string {
	var sb strings.Builder
	sb.WriteString("0x")
	for _, b := range e {
		str := strconv.FormatInt(int64(b), 16)
		if len(str) == 2 {
			sb.WriteString(str)
		} else {
			sb.WriteString("0")
			sb.WriteString(str)
		}
	}
	return sb.String()
}
