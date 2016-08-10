package nv

import (
	"encoding/binary"
	"io"
	"reflect"
	"time"

	"github.com/cerana/cerana/pkg/errors"
	xdr "github.com/davecgh/go-xdr/xdr2"
)

// XDRDecoder is a Decoder for XDR encoding.
type XDRDecoder struct {
	*xdr.Decoder
	r    io.ReadSeeker
	pair pair
}

// NewXDRDecoder creates a new XDRDecoder.
func NewXDRDecoder(r io.ReadSeeker) *XDRDecoder {
	return &XDRDecoder{Decoder: xdr.NewDecoder(r), r: r}
}

// Decode decodes data into a supplied target.
// Note: care should be taken when decoding into a `map[string]interface{}` as
// bytes/uint8s (and their array forms) can not be distinguished and will be
// treated as uint8/[]uint8.
func (d *XDRDecoder) Decode(target interface{}) error {
	// Validate data encoding
	codec, endianness, err := decodePreamble(d.r, binary.BigEndian)
	if err != nil {
		return err
	} else if codec != xdrCodec {
		return errors.Newv("invalid encoding", map[string]interface{}{"codec": codec})
	} else if endianness != littleEndian {
		return errors.Newv("invalid endianess", map[string]interface{}{"endianness": endianness})
	}

	// Validate target
	targetV := reflect.ValueOf(target)
	if targetV.Kind() != reflect.Ptr {
		return errors.Newv("cannot decode into non-pointer", map[string]interface{}{"type": reflect.TypeOf(targetV).String()})
	}
	if targetV.IsNil() {
		return errors.New("cannot decode into nil")
	}

	return decodeList(d, reflect.Indirect(targetV))
}

func (d *XDRDecoder) header() (header, error) {
	var h header
	err := binary.Read(d.r, binary.BigEndian, &h)
	return h, errors.Wrap(err, "failed to decode header")
}

func (d *XDRDecoder) meta() (string, dataType, error) {
	_, err := xdr.Unmarshal(d.r, &d.pair)
	return d.pair.Name, d.pair.Type, errors.Wrap(err, "failed to decode nvpair metadata")
}

func (d *XDRDecoder) skip() error {
	_, err := d.r.Seek(int64(d.pair.EncodedSize-uint32(d.pair.headerSize())), 1)
	return errors.Wrap(err, "failed to skip")
}

func (d *XDRDecoder) isEnd() (bool, error) {
	var end uint64
	err := binary.Read(d.r, binary.BigEndian, &end)
	if err != nil {
		return false, errors.Wrap(err, "failed to check for end")
	}
	if end == 0 {
		return true, nil
	}
	_, err = d.r.Seek(-8, 1)
	return false, errors.Wrap(err, "failed to seek back")
}

func (d *XDRDecoder) value(targetType reflect.Type) (reflect.Value, fieldSetFunc, error) {
	var val reflect.Value
	var fsf fieldSetFunc
	var err error

	var v interface{}
	switch d.pair.Type {
	case _boolean:
		v := Boolean(true)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _booleanValue:
		v, err = d.decodeBool()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetBool(v.(bool))
		}
	case _byte:
		v, err = d.decodeByte()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _int8:
		v, err = d.decodeInt8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int8)))
		}
	case _int16:
		v, err = d.decodeInt16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int16)))
		}
	case _int32:
		v, err = d.decodeInt32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int32)))
		}
	case _int64:
		v, err = d.decodeInt64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(v.(int64))
		}
	case _uint8:
		v, err = d.decodeUint8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _uint16:
		v, err = d.decodeUint16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint16)))
		}
	case _uint32:
		v, err = d.decodeUint32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint32)))
		}
	case _uint64:
		v, err = d.decodeUint64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint64)))
		}
	case _hrtime:
		v, err = d.decodeHRTime()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(time.Duration)))
		}
	case _double:
		v, err = d.decodeFloat64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetFloat(v.(float64))
		}
	case _booleanArray:
		v, err = d.decodeBoolArray()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _byteArray:
		if _, err = d.r.Seek(-4, 1); err == nil {
			v, err = d.decodeByteArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.SetBytes(v.([]byte))
			}
		}
	case _int8Array:
		v, err = d.decodeInt8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int16Array:
		v, err = d.decodeInt16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int32Array:
		v, err = d.decodeInt32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int64Array:
		v, err = d.decodeInt64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint8Array:
		v, err = d.decodeUint8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint16Array:
		v, err = d.decodeUint16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint32Array:
		v, err = d.decodeUint32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint64Array:
		v, err = d.decodeUint64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _string:
		v, err = d.decodeString()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetString(v.(string))
		}
	case _stringArray:
		if _, err = d.r.Seek(-4, 1); err == nil {
			v, err = d.decodeStringArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.Set(val)
			}
		}
	case _nvlist:
		val = reflect.Indirect(reflect.New(targetType))
		err = decodeList(d, val)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _nvlistArray:
		if targetType.Kind() == reflect.Interface {
			targetType = reflect.TypeOf([]map[string]interface{}{})
		}

		nelems := int(d.pair.NElements)
		val = reflect.MakeSlice(targetType, 0, nelems)
		for i := 0; i < nelems; i++ {
			elem := reflect.Indirect(reflect.New(targetType.Elem()))
			err = decodeList(d, elem)
			if err != nil {
				break
			}
			val = reflect.Append(val, elem)
		}
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	default:
		err = errors.Newv("unknown type", map[string]interface{}{"type": d.pair.Type})
	}
	return val, fsf, err
}
