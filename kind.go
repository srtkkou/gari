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

const (
	kindBool   = "Bool"
	kindString = "String"
	kindTime   = "Time"
)

var (
	// msgpackエンコードエラー
	ErrMsgpackEncode = errors.New("orm.ErrMsgpackEncode")
	// msgpackデコードエラー
	ErrMsgpackDecode = errors.New("orm.ErrMsgpackDecode")
	// 型の不一致
	ErrTypeMismatch = errors.New("orm.ErrTypeMismatch")
)

// 真偽値をエンコードする。
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
				errs.WithContext("bool", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullBool:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err := enc.EncodeBool(tv.Bool); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("bool", tv.Bool))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrTypeMismatch, errs.WithContext("value", v))
		return []byte{}, err
	}
}

// 文字列をエンコードする。
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
				errs.WithContext("string", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullString:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeString(tv.String); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("string", tv.String))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrTypeMismatch, errs.WithContext("value", v))
		return []byte{}, err
	}
}

// 時刻をエンコードする。
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
				errs.WithContext("time", tv))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	case sql.NullTime:
		if !tv.Valid {
			return msgpackNil(), nil
		}
		if err = enc.EncodeTime(tv.Time); err != nil {
			err = errs.Wrap(ErrMsgpackEncode, errs.WithCause(err),
				errs.WithContext("time", tv.Time))
			return []byte{}, err
		}
		return buf.Bytes(), nil
	default:
		err = errs.Wrap(ErrTypeMismatch, errs.WithContext("value", v))
		return []byte{}, err
	}
}

/*
// 整数値をエンコードする。
func (k kind) encodeInt64(value any) string {
	if k != kindInt {
		return lib.JSON_NULL
	}
	num := int64(0)
	switch tv := value.(type) {
	case sql.NullInt64:
		if !tv.Valid {
			return lib.JSON_NULL
		}
		num = tv.Int64
	case sql.NullInt32:
		if !tv.Valid {
			return lib.JSON_NULL
		}
		num = int64(tv.Int32)
	case sql.NullInt16:
		if !tv.Valid {
			return lib.JSON_NULL
		}
		num = int64(tv.Int16)
	case sql.NullFloat64:
		if !tv.Valid {
			return lib.JSON_NULL
		}
		num = int64(tv.Float64)
	case int:
		num = int64(tv)
	case int64:
		num = tv
	case int32:
		num = int64(tv)
	case int16:
		num = int64(tv)
	case uint:
		num = int64(tv)
	case uint64:
		num = int64(tv)
	case uint32:
		num = int64(tv)
	case uint16:
		num = int64(tv)
	case float64:
		num = int64(tv)
	case float32:
		num = int64(tv)
	default:
		return lib.JSON_NULL
	}
	return strconv.FormatInt(num, 10)
}
*/

// 真偽値をデコードする。
func decodeNullBool(blob []byte) (sql.NullBool, error) {
	nb := sql.NullBool{Valid: false}
	// nil値の復元
	if slices.Equal(blob, msgpackNil()) {
		return nb, nil
	}
	// 真偽値の復元
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	b, err := dec.DecodeBool()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("bytes", blob))
		return nb, err
	}
	nb.Valid = true
	nb.Bool = b
	return nb, nil
}

// 文字列をデコードする。
func decodeNullString(blob []byte) (sql.NullString, error) {
	ns := sql.NullString{Valid: false}
	// nil値の復元
	if slices.Equal(blob, msgpackNil()) {
		return ns, nil
	}
	// 文字列の復元
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	str, err := dec.DecodeString()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("bytes", blob))
		return ns, err
	}
	ns.Valid = true
	ns.String = str
	return ns, nil
}

// 時刻をデコードする。
func decodeNullTime(blob []byte) (sql.NullTime, error) {
	nt := sql.NullTime{Valid: false}
	// nil値の復元
	if slices.Equal(blob, msgpackNil()) {
		return nt, nil
	}
	// 時間の復元
	r := bytes.NewReader(blob)
	dec := msgpack.NewDecoder(r)
	t, err := dec.DecodeTime()
	if err != nil {
		err = errs.Wrap(ErrMsgpackDecode, errs.WithCause(err),
			errs.WithContext("bytes", blob))
		return nt, err
	}
	nt.Valid = true
	nt.Time = t.UTC()
	return nt, nil
}

/*
// 整数値をデコードする。
func (k kind) decodeInt64(str string) int64 {
	if k == kindInt && str != lib.JSON_NULL {
		num, err := strconv.ParseInt(str, 10, 64)
		if err == nil {
			return num
		}
	}
	return 0
}
*/

// msgpack形式のnil
func msgpackNil() []byte {
	return []byte{0xc0}
}
