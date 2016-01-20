package nv

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

type XDRDecoder struct {
	*xdr.Decoder
	r    io.ReadSeeker
	pair pair
}

func NewXDRDecoder(r io.ReadSeeker) *XDRDecoder {
	return &XDRDecoder{Decoder: xdr.NewDecoder(r), r: r}
}

// Decode
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
	return decHeader(d.r)
}

func decHeader(r io.ReadSeeker) (header, error) {
	var h header
	err := binary.Read(r, binary.BigEndian, &h)
	return h, err
}

func (d *XDRDecoder) meta() (string, dataType, error) {
	err := decMeta(d.r, &d.pair)
	return d.pair.Name, d.pair.Type, err
}

func decMeta(r io.ReadSeeker, pair *pair) error {
	_, err := xdr.Unmarshal(r, pair)
	return err
}

func (d *XDRDecoder) skip() error {
	return skip(d.r, d.pair)
}

func skip(r io.ReadSeeker, pair pair) error {
	_, err := r.Seek(int64(pair.EncodedSize-uint32(pair.headerSize())), 1)
	return err
}

func (d *XDRDecoder) isEnd() (bool, error) {
	return isEnd(d.r)
}

func isEnd(r io.ReadSeeker) (bool, error) {
	var end uint64
	err := binary.Read(r, binary.BigEndian, &end)
	if err != nil {
		return false, err
	}
	if end == 0 {
		return true, nil
	}
	_, err = r.Seek(-8, 1)
	return false, err
}

func (d *XDRDecoder) value(targetType reflect.Type) (reflect.Value, fieldSetFunc, error) {
	var val reflect.Value
	var fsf fieldSetFunc
	err := fmt.Errorf("unknown type: %v", d.pair.Type)

	var v interface{}
	switch d.pair.Type {
	case _BOOLEAN:
		err = nil
		v := Boolean(true)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _BOOLEAN_VALUE:
		v, err = d.decodeBool()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetBool(v.(bool))
		}
	case _BYTE:
		v, err = d.decodeByte()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _INT8:
		v, err = d.decodeInt8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int8)))
		}
	case _INT16:
		v, err = d.decodeInt16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int16)))
		}
	case _INT32:
		v, err = d.decodeInt32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int32)))
		}
	case _INT64:
		v, err = d.decodeInt64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(v.(int64))
		}
	case _UINT8:
		v, err = d.decodeUint8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _UINT16:
		v, err = d.decodeUint16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint16)))
		}
	case _UINT32:
		v, err = d.decodeUint32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint32)))
		}
	case _UINT64:
		v, err = d.decodeUint64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint64)))
		}
	case _HRTIME:
		v, err = d.decodeHRTime()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(time.Duration)))
		}
	case _DOUBLE:
		v, err = d.decodeFloat64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetFloat(v.(float64))
		}
	case _BOOLEAN_ARRAY:
		v, err = d.decodeBoolArray()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _BYTE_ARRAY:
		if _, err = d.r.Seek(-4, 1); err == nil {
			v, err = d.decodeByteArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.SetBytes(v.([]byte))
			}
		}
	case _INT8_ARRAY:
		v, err = d.decodeInt8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT16_ARRAY:
		v, err = d.decodeInt16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT32_ARRAY:
		v, err = d.decodeInt32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT64_ARRAY:
		v, err = d.decodeInt64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT8_ARRAY:
		v, err = d.decodeUint8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT16_ARRAY:
		v, err = d.decodeUint16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT32_ARRAY:
		v, err = d.decodeUint32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT64_ARRAY:
		v, err = d.decodeUint64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _STRING:
		v, err = d.decodeString()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetString(v.(string))
		}
	case _STRING_ARRAY:
		if _, err = d.r.Seek(-4, 1); err == nil {
			v, err = d.decodeStringArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.Set(val)
			}
		}
	case _NVLIST:
		val = reflect.Indirect(reflect.New(targetType))
		err = decodeList(d, val)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _NVLIST_ARRAY:
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
