package gari

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"math"
	"slices"
	"time"

	"github.com/goark/errs"
	"github.com/vmihailenco/msgpack/v5"
)

const (
	fixArrayByte = byte(0b1001_0000)
	array16Byte  = byte(0xdc)
	array32Byte  = byte(0xdd)
	fixMapByte   = byte(0b1000_0000)
	map16Byte    = byte(0xde)
	map32Byte    = byte(0xdf)
)

var (
	ErrMsgpackEncode    = errors.New("gari.msgpack.ErrMsgpackEncode")
	ErrMsgpackDecode    = errors.New("gari.ErrMsgpackDecode")
	ErrMsgpackInputType = errors.New("gari.ErrMsgpackInputType")
	ErrMsgpackArraySize = errors.New("gari.ErrMsgpackArraySize")
	ErrMsgpackMapSize   = errors.New("gari.ErrMsgpackMapSize")
	ErrMsgpackMapToType = errors.New("gari.ErrMsgpackMapToType")
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
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullBool:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err := enc.EncodeBool(tv.Bool); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv.Bool))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrMsgpackInputType,
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
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullString:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeString(tv.String); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv.String))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrMsgpackInputType,
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
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullTime:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeTime(tv.Time); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("input", tv.Time))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrMsgpackInputType,
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
		err := errs.Wrap(ErrMsgpackInputType,
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
		err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
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

// Decode to map.
func decodeToMap(blob []byte) (map[string]any, error) {
	// Decode msgpack bytes as map.
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	m, err := dec.DecodeMap()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("input", blob))
		return nil, err
	}
	// Convert map values.
	for k, v := range m {
		switch tv := v.(type) {
		case map[string]any:
			// Try to convert to nullable sql types.
			if ns, err := mapToNullString(tv); err == nil {
				m[k] = ns
			} else if nt, err := mapToNullTime(tv); err == nil {
				m[k] = nt
			}
		}
	}
	return m, nil
}

// NULL in msgpack format.
func msgpackNil() []byte {
	return []byte{0xc0}
}

// Array header.
func msgpackArrayHeader(size int) ([]byte, error) {
	var b bytes.Buffer
	if size <= 15 { // fixarray
		headByte := fixArrayByte | byte(size)
		b.WriteByte(headByte)
	} else if size <= math.MaxUint16 { // array16
		b.WriteByte(array16Byte)
		binary.Write(&b, binary.BigEndian, uint16(size))
	} else if size <= math.MaxUint32 { // array32
		b.WriteByte(array32Byte)
		binary.Write(&b, binary.BigEndian, uint32(size))
	} else {
		err := errs.Wrap(ErrMsgpackArraySize,
			errs.WithContext("size", size))
		return []byte{}, err
	}
	return b.Bytes(), nil
}

// Map header.
func msgpackMapHeader(size int) ([]byte, error) {
	var b bytes.Buffer
	if size <= 15 { // fixmap
		headByte := fixMapByte | byte(size)
		b.WriteByte(headByte)
	} else if size <= math.MaxUint16 { // map16
		b.WriteByte(map16Byte)
		binary.Write(&b, binary.BigEndian, uint16(size))
	} else if size <= math.MaxUint32 { // map32
		b.WriteByte(map32Byte)
		binary.Write(&b, binary.BigEndian, uint32(size))
	} else {
		err := errs.Wrap(ErrMsgpackMapSize,
			errs.WithContext("size", size))
		return []byte{}, err
	}
	return b.Bytes(), nil
}

func mapToNullString(m map[string]any) (ns sql.NullString, err error) {
	// Decode m["Valid"] value.
	ns.Valid, err = mapToNullable(m)
	if err != nil {
		return ns, err
	}
	// Decode m["String"] value.
	v, ok := m["String"]
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("String", v))
		return ns, err
	}
	ns.String, ok = v.(string)
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("String", v))
		return ns, err
	}
	return ns, nil
}

func mapToNullTime(m map[string]any) (nt sql.NullTime, err error) {
	// Decode m["Valid"] value.
	nt.Valid, err = mapToNullable(m)
	if err != nil {
		return nt, err
	}
	// Decode m["Time"] value.
	v, ok := m["Time"]
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("Time", v))
		return nt, err
	}
	nt.Time, ok = v.(time.Time)
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("Time", v))
		return nt, err
	}
	return nt, nil
}

func mapToNullable(m map[string]any) (valid bool, err error) {
	// Decode m["Valid"] value.
	v, ok := m["Valid"]
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("Valid", v))
		return valid, err
	}
	valid, ok = v.(bool)
	if !ok {
		err = errs.Wrap(ErrMsgpackMapToType,
			errs.WithContext("Valid", v))
		return valid, err
	}
	return valid, nil
}
