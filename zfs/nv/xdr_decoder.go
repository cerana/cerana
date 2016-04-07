package nv

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"

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
		return fmt.Errorf("invalid encoding: %v", codec)
	} else if endianness != littleEndian {
		return fmt.Errorf("invalid endianess: %v", endianness)
	}

	// Validate target
	targetV := reflect.ValueOf(target)
	if targetV.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode into non-pointer: %v", reflect.TypeOf(targetV).String())
	}
	if targetV.IsNil() {
		return fmt.Errorf("cannot decode into nil")
	}

	return decodeList(d, reflect.Indirect(targetV))
}

func (d *XDRDecoder) header() (header, error) {
	var h header
	err := binary.Read(d.r, binary.BigEndian, &h)
	return h, err
}

func (d *XDRDecoder) meta() (string, dataType, error) {
	_, err := xdr.Unmarshal(d.r, &d.pair)
	return d.pair.Name, d.pair.Type, err
}

func (d *XDRDecoder) skip() error {
	_, err := d.r.Seek(int64(d.pair.EncodedSize-uint32(d.pair.headerSize())), 1)
	return err
}

func (d *XDRDecoder) isEnd() (bool, error) {
	var end uint64
	err := binary.Read(d.r, binary.BigEndian, &end)
	if err != nil {
		return false, err
	}
	if end == 0 {
		return true, nil
	}
	_, err = d.r.Seek(-8, 1)
	return false, err
}

func (d *XDRDecoder) value(targetType reflect.Type) (reflect.Value, fieldSetFunc, error) {
	var val reflect.Value
	var fsf fieldSetFunc
	err := fmt.Errorf("unknown type: %v", d.pair.Type)

	var v interface{}
	switch d.pair.Type {
	case _boolean:
		err = nil
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
	}
	return val, fsf, err
}
