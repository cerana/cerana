package nv

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"sort"
)

var (
	types = map[reflect.Kind]dataType{
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

	sliceTypes = map[reflect.Kind]dataType{
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
)

func validValue(v reflect.Value) error {
	if !v.IsValid() {
		return fmt.Errorf("type '%s' is invalid", v.Kind().String())
	}

	return nil
}

func encodePreamble(w io.Writer, codec codec, order endianness) error {
	return binary.Write(w, binary.BigEndian, encoding{Encoding: codec, Endianess: order})
}

func encodeList(w io.Writer, v reflect.Value) error {
	if err := encHeader(w, header{Flag: _UNIQUE_NAME}); err != nil {
		return err
	}

	v = deref(v)
	switch v.Kind() {
	case reflect.Struct:
		if err := encodeStruct(w, v); err != nil {
			return err
		}
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
	default:
		return fmt.Errorf("invalid type '%s', must be a struct", v.Kind().String())
	}

	return encFooter(w)
}

func encodeStruct(w io.Writer, v reflect.Value) error {
	var err error

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

	return err
}

func encodeItem(w io.Writer, name string, tags []string, field reflect.Value) error {
	var tagType dataType
	if len(tags) > 1 {
		if tags[1] == "byte" {
			tagType = _BYTE
		} else if tags[1] == "uint8" {
			tagType = _UINT8
		}
	}

	field = deref(field)
	dtype, ok := types[field.Kind()]

	switch field.Kind() {
	case reflect.Bool:
		if field.Type().Name() == "Boolean" {
			dtype = _BOOLEAN
		}
	case reflect.Interface:
		return encodeItem(w, name, tags, reflect.ValueOf(field.Interface()))
	case reflect.Slice, reflect.Array:
		dtype, ok = sliceTypes[field.Type().Elem().Kind()]
		switch tagType {
		case _BYTE:
			dtype = _BYTE_ARRAY
		case _UINT8:
			dtype = _UINT8_ARRAY
		}
	case reflect.Int64:
		if field.Type().String() == "time.Duration" {
			dtype = _HRTIME
		}
	case reflect.Uint8:
		switch tagType {
		case _BYTE:
			dtype = _BYTE
		case _UINT8:
			dtype = _UINT8
		}
	}

	if !ok {
		return fmt.Errorf("unknown type: %v", field.Kind())
	}

	return xdrEncode(w, name, dtype, field.Interface())
}
