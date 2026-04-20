package gari

import (
	"bytes"
	"database/sql"
	"errors"
	"slices"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

var (
	// msgpack encode error.
	ErrEncode = errors.New("orm.ErrEncode")
	// msgpack decode error.
	ErrDecode = errors.New("orm.ErrDecode")
	// msgpack encode input type error.
	ErrEncodeInputType = errors.New("orm.ErrEncodeInputType")
)

// Encode NULL or bool value.
func encodeNullBool(v any) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	switch tv := v.(type) {
	case nil:
		return msgpackNil(), nil
	case bool:
		if err := enc.EncodeBool(tv); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullBool:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err := enc.EncodeBool(tv.Bool); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv.Bool))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrEncodeInputType,
			errs.WithContext("input", v))
		return []byte{}, err
	}
}

// Encode NULL or string value.
func encodeNullString(v any) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	switch tv := v.(type) {
	case nil:
		return msgpackNil(), nil
	case string:
		if err = enc.EncodeString(tv); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullString:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeString(tv.String); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv.String))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrEncodeInputType,
			errs.WithContext("input", v))
		return []byte{}, err
	}
}

// Encode NULL or time.Time value.
func encodeNullTime(v any) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	switch tv := v.(type) {
	case nil:
		return msgpackNil(), nil
	case time.Time:
		if err = enc.EncodeTime(tv); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullTime:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeTime(tv.Time); err != nil {
			err = errs.Wrap(ErrEncode, errs.WithCause(err),
				errs.WithContext("input", tv.Time))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrEncodeInputType,
			errs.WithContext("input", v))
		return []byte{}, err
	}
}

// Encode NULL or int value.
func encodeNullInt64(v any) ([]byte, error) {
	switch tv := v.(type) {
	case nil:
		return msgpackNil(), nil
	case sql.NullInt16:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		return encodeInt64(int64(tv.Int16))
	case sql.NullInt32:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		return encodeInt64(int64(tv.Int32))
	case sql.NullInt64:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		return encodeInt64(tv.Int64)
	case int:
		return encodeInt64(int64(tv))
	case int16:
		return encodeInt64(int64(tv))
	case int32:
		return encodeInt64(int64(tv))
	case int64:
		return encodeInt64(tv)
	default:
		err := errs.Wrap(ErrEncodeInputType,
			errs.WithContext("input", v))
		return []byte{}, err
	}
}

// Encode int64. (Internal use only)
func encodeInt64(num int64) ([]byte, error) {
	var err error
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err = enc.EncodeInt64(num); err != nil {
		err = errs.Wrap(ErrEncode, errs.WithCause(err),
			errs.WithContext("input", num))
		return []byte{}, err
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
		err = errs.Wrap(ErrDecode, errs.WithCause(err),
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
		err = errs.Wrap(ErrDecode, errs.WithCause(err),
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
		err = errs.Wrap(ErrDecode, errs.WithCause(err),
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
		err = errs.Wrap(ErrDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return ni, err
	}
	ni.Valid = true
	ni.Int64 = num
	return ni, nil
}

// NULL in msgpack format.
func msgpackNil() []byte {
	return []byte{0xc0}
}
