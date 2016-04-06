package nv

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"time"
)

var empty8 = [8]byte{}

type embeddedHeader struct {
	header
	Priv uint64
	Flag uint32
	Pad  int32
}

type nativeMeta struct {
	Size      uint32
	NameLen   uint32
	NElements uint32
	Type      dataType
}

type nativePair struct {
	nativeMeta
	Name string
	Data interface{}
}

type NativeDecoder struct {
	r           io.ReadSeeker
	pair        nativePair
	size        uint32
	embedded    bool
	savedHeader header
}

func NewNativeDecoder(r io.ReadSeeker) *NativeDecoder {
	return &NativeDecoder{r: r}
}

// Decode
// Note: care should be taken when decoding into a `map[string]interface{}` as
// bytes/uint8s (and their array forms) can not be distinguished and will be
// treated as uint8/[]uint8.
func (d *NativeDecoder) Decode(target interface{}) (err error) {
	// Validate data encoding
	codec, endianness, err := decodePreamble(d.r, binary.BigEndian)
	if err != nil {
		return err
	} else if codec != nativeCodec {
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
		return fmt.Errorf("cannot decode into nil pointer")
	}

	return decodeList(d, reflect.Indirect(targetV))
}

func (d *NativeDecoder) header() (header, error) {
	var h header
	if !d.embedded {
		return h, binary.Read(d.r, binary.LittleEndian, &h)
	}
	return d.savedHeader, nil
}

func decodeNativeEmbeddedListHeader(r io.ReadSeeker) (header, error) {
	var h embeddedHeader
	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return header{}, err
	}
	if h.Priv != 0 || h.Flag != 0 || h.Pad != 0 {
		return header{}, errors.New("invalid embedded header")
	}
	return h.header, nil
}

func (d *NativeDecoder) meta() (string, dataType, error) {
	err := binary.Read(d.r, binary.LittleEndian, &d.pair.nativeMeta)
	if err != nil {
		return "", 0, err
	}

	len := uint32(align8(int(d.pair.NameLen)))
	buf := make([]byte, len)
	if _, err := d.r.Read(buf); err != nil {
		return "", 0, err
	}
	if len == 0 {
		return "", 0, errors.New("wtf")
	}

	d.pair.Name = string(buf[:d.pair.NameLen-1])

	d.size = d.pair.Size - 16 - len
	return d.pair.Name, d.pair.Type, err
}

func (d *NativeDecoder) skip() error {
	_, err := d.r.Seek(int64(align8(int(d.size))), 1)
	return err
}

func (d *NativeDecoder) isEnd() (bool, error) {
	var end uint32
	err := binary.Read(d.r, binary.LittleEndian, &end)
	if err != nil {
		return false, err
	}
	if end == 0 {
		return true, nil
	}
	_, err = d.r.Seek(-4, 1)
	return false, err
}

func (d *NativeDecoder) value(targetType reflect.Type) (reflect.Value, fieldSetFunc, error) {
	var val reflect.Value
	var fsf fieldSetFunc
	err := fmt.Errorf("unknown type: %v", d.pair.Type)

	var seek int64

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
		seek = 4
		v, err = d.decodeBool()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetBool(v.(bool))
		}
	case _byte:
		seek = 7
		v, err = d.decodeByte()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _int8:
		seek = 7
		v, err = d.decodeInt8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int8)))
		}
	case _int16:
		seek = 6
		v, err = d.decodeInt16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int16)))
		}
	case _int32:
		v, err = d.decodeInt32()
		seek = 4
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
		seek = 4
		v, err = d.decodeUint8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _uint16:
		seek = 6
		v, err = d.decodeUint16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint16)))
		}
	case _uint32:
		seek = 4
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
		used := 4 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeBoolArray(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _byteArray:
		used := 1 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeByteArray(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int8Array:
		used := 1 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeInt8Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int16Array:
		used := 2 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeInt16Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int32Array:
		used := 4 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeInt32Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _int64Array:
		v, err = d.decodeInt64Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint8Array:
		used := 1 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeUint8Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint16Array:
		used := 2 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeUint16Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint32Array:
		used := 4 * d.pair.NElements
		seek = int64(d.size - used)
		v, err = d.decodeUint32Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _uint64Array:
		v, err = d.decodeUint64Array(d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _string:
		v, err = d.decodeString(int(d.size))
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetString(v.(string))
		}
	case _stringArray:
		skip := 8 * int64(d.pair.NElements)
		if _, err = d.r.Seek(skip, 1); err != nil {
			break
		}

		len := int64(d.size) - skip
		v, err = d.decodeStringArray(len, d.pair.NElements)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _nvlist:
		val = reflect.Indirect(reflect.New(targetType))

		dec := NewNativeDecoder(d.r)
		dec.embedded = true
		dec.savedHeader, err = decodeNativeEmbeddedListHeader(d.r)
		if err != nil {
			break
		}
		err = decodeList(dec, val)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _nvlistArray:
		if targetType.Kind() == reflect.Interface {
			targetType = reflect.TypeOf([]map[string]interface{}{})
		}

		if _, err = d.r.Seek(int64(8*d.pair.NElements), 1); err != nil {
			break
		}

		headers := make([]header, d.pair.NElements)
		for i := uint32(0); i < d.pair.NElements; i++ {
			headers[i], err = decodeNativeEmbeddedListHeader(d.r)
			if err != nil {
				break
			}
		}
		if err != nil {
			break
		}

		val = reflect.MakeSlice(targetType, 0, int(d.pair.NElements))
		for i := uint32(0); i < d.pair.NElements; i++ {
			elem := reflect.Indirect(reflect.New(targetType.Elem()))
			dec := NewNativeDecoder(d.r)
			dec.embedded = true
			dec.savedHeader = headers[i]
			if err = decodeList(dec, elem); err != nil {
				break
			}
			val = reflect.Append(val, elem)
		}
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	}

	if err != nil {
		return reflect.Value{}, nil, err
	}

	_, err = d.r.Read(empty8[:seek])
	if err != nil {
		return reflect.Value{}, nil, err
	}
	return val, fsf, nil
}
