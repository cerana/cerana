package nv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

func Encode(i interface{}) ([]byte, error) {
	if i == nil {
		return nil, errors.New("can not encode a nil pointer")
	}

	v := reflect.ValueOf(i)
	if !v.IsValid() {
		return nil, fmt.Errorf("type '%s' is invalid", v.Kind().String())
	}

	var err error
	buff := bytes.NewBuffer(nil)
	if err = binary.Write(buff, binary.BigEndian, encoding{Encoding: 1, Endianess: 1}); err != nil {
		return nil, err
	}

	if err = encodeList(buff, v); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func encodeList(w io.Writer, v reflect.Value) error {
	var err error
	if err = binary.Write(w, binary.BigEndian, header{Flag: _UNIQUE_NAME}); err != nil {
		return err
	}

	v = deref(v)
	switch v.Kind() {
	case reflect.Struct:
		_, err = encodeStruct(v, w)
	case reflect.Map:
		keys := make([]string, len(v.MapKeys()))
		for i, k := range v.MapKeys() {
			keys[i] = k.Interface().(string)
		}
		sort.Strings(keys)

		for _, name := range keys {
			v := v.MapIndex(reflect.ValueOf(name))
			if err := encodeItem(w, name, nil, v); err != nil {
				return err
			}
		}
		err = binary.Write(w, binary.BigEndian, uint64(0))
	default:
		return fmt.Errorf("invalid type '%s', must be a struct", v.Kind().String())
	}

	return err
}

func encodeStruct(v reflect.Value, w io.Writer) (int, error) {
	var err error
	size := 0

	forEachField(v, func(i int, field reflect.Value) bool {
		// Skip fields that can't be set (e.g. unexported)
		if !field.CanSet() {
			return true
		}
		name := v.Type().Field(i).Name
		tags := getTags(i, v)
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}

		if err = encodeItem(w, name, tags, field); err != nil {
			return false
		}
		return true
	})

	if err != nil {
		return 0, err
	}

	if err = binary.Write(w, binary.BigEndian, uint64(0)); err != nil {
		return 0, err
	}
	return size + 8, nil
}

func encodeItem(w io.Writer, name string, tags []string, field reflect.Value) error {
	field = deref(field)
	var types = map[reflect.Kind]dataType{
		reflect.Bool:    _BOOLEAN_VALUE,
		reflect.Float32: _DOUBLE,
		reflect.Float64: _DOUBLE,
		reflect.Int16:   _INT16,
		reflect.Int32:   _INT32,
		reflect.Int64:   _INT64,
		reflect.Int8:    _INT8,
		reflect.Int:     _INT32,
		reflect.Map:     _NVLIST,
		reflect.String:  _STRING,
		reflect.Struct:  _NVLIST,
		reflect.Uint16:  _UINT16,
		reflect.Uint32:  _UINT32,
		reflect.Uint64:  _UINT64,
		reflect.Uint8:   _UINT8,
		reflect.Uint:    _UINT32,
	}

	var sliceTypes = map[reflect.Kind]dataType{
		reflect.Bool:   _BOOLEAN_ARRAY,
		reflect.Int16:  _INT16_ARRAY,
		reflect.Int32:  _INT32_ARRAY,
		reflect.Int64:  _INT64_ARRAY,
		reflect.Int8:   _INT8_ARRAY,
		reflect.Int:    _INT32_ARRAY,
		reflect.Map:    _NVLIST_ARRAY,
		reflect.String: _STRING_ARRAY,
		reflect.Struct: _NVLIST_ARRAY,
		reflect.Uint16: _UINT16_ARRAY,
		reflect.Uint32: _UINT32_ARRAY,
		reflect.Uint64: _UINT64_ARRAY,
		reflect.Uint8:  _UINT8_ARRAY,
		reflect.Uint:   _UINT32_ARRAY,
	}
	var tagType dataType
	if len(tags) > 1 {
		if tags[1] == "byte" {
			tagType = _BYTE
		} else if tags[1] == "uint8" {
			tagType = _UINT8
		}
	}

	p := pair{
		Name:      name,
		NElements: 1,
	}

	var ok bool
	p.Type, ok = types[field.Kind()]

	switch field.Kind() {
	case reflect.Bool:
		if field.Type().Name() == "Boolean" {
			p.Type = _BOOLEAN
		}
	case reflect.Interface:
		return encodeItem(w, name, tags, reflect.ValueOf(field.Interface()))
	case reflect.Slice, reflect.Array:
		p.Type, ok = sliceTypes[field.Type().Elem().Kind()]
		switch tagType {
		case _BYTE:
			p.Type = _BYTE_ARRAY
		case _UINT8:
			p.Type = _UINT8_ARRAY
		}
	case reflect.Int64:
		if field.Type().String() == "time.Duration" {
			p.Type = _HRTIME
		}
	case reflect.Uint8:
		switch tagType {
		case _BYTE:
			p.Type = _BYTE
		case _UINT8:
			p.Type = _UINT8
		}
	}

	if !ok {
		return fmt.Errorf("unknown type: %v", field.Kind())
	}

	p.data = field.Interface()
	value := p.data
	vbuf := &bytes.Buffer{}
	switch p.Type {
	case _BOOLEAN:
		p.NElements = 0
	case _BYTE:
		value = int8(value.(uint8))
	case _UINT8:
		value = int(int8(value.(uint8)))
	case _BYTE_ARRAY:
		p.NElements = uint32(len(value.([]byte)))
		n := int(p.NElements)
		arrType := reflect.ArrayOf(n, reflect.TypeOf(byte(0)))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]byte) {
			arr.Index(i).SetUint(uint64(b))
		}
		value = arr.Interface()
	case _BOOLEAN_ARRAY:
		p.NElements = uint32(len(value.([]bool)))
	case _INT8_ARRAY:
		p.NElements = uint32(len(value.([]int8)))
	case _INT16_ARRAY:
		p.NElements = uint32(len(value.([]int16)))
	case _INT32_ARRAY:
		p.NElements = uint32(len(value.([]int32)))
	case _INT64_ARRAY:
		p.NElements = uint32(len(value.([]int64)))
	case _UINT8_ARRAY:
		// this one is weird since UINT8s are encoded as char
		// aka int32s... :(
		p.NElements = uint32(len(value.([]uint8)))
		n := int(p.NElements)
		sliceType := reflect.SliceOf(reflect.TypeOf(int32(0)))
		slice := reflect.MakeSlice(sliceType, n, n)
		for i, b := range value.([]uint8) {
			slice.Index(i).SetInt(int64(int8(b)))
		}
		value = slice.Interface()
	case _UINT16_ARRAY:
		p.NElements = uint32(len(value.([]uint16)))
	case _UINT32_ARRAY:
		p.NElements = uint32(len(value.([]uint32)))
	case _UINT64_ARRAY:
		p.NElements = uint32(len(value.([]uint64)))
	case _STRING_ARRAY:
		p.NElements = uint32(len(value.([]string)))
		arrType := reflect.ArrayOf(int(p.NElements), reflect.TypeOf(""))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]string) {
			arr.Index(i).SetString(b)
		}
		value = arr.Interface()
	case _NVLIST:
		if err := encodeList(vbuf, reflect.ValueOf(value)); err != nil {
			return err
		}
		p.data = vbuf.Bytes()
	case _NVLIST_ARRAY:
		p.NElements = uint32(len(value.([]map[string]interface{})))
		for _, l := range value.([]map[string]interface{}) {
			if err := encodeList(vbuf, reflect.ValueOf(l)); err != nil {
				return err
			}
		}
		p.data = vbuf.Bytes()
	}

	if vbuf.Len() == 0 && p.Type != _BOOLEAN {
		_, err := xdr.NewEncoder(vbuf).Encode(value)
		if err != nil {
			return err
		}
	}

	p.EncodedSize = uint32(p.encodedSize())
	p.DecodedSize = uint32(p.decodedSize())

	pbuf := &bytes.Buffer{}
	_, err := xdr.NewEncoder(pbuf).Encode(p)
	if err != nil {
		return err
	}

	_, err = pbuf.WriteTo(w)
	if err != nil {
		return err
	}
	_, err = vbuf.WriteTo(w)
	if err != nil {
		return err
	}

	return nil
}
