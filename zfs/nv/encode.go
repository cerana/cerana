package nv

import (
	"encoding/binary"
	"io"
	"reflect"
	"sort"

	"github.com/cerana/cerana/pkg/errors"
)

var (
	types = map[reflect.Kind]dataType{
		reflect.Bool:    _booleanValue,
		reflect.Float32: _double,
		reflect.Float64: _double,
		reflect.Int16:   _int16,
		reflect.Int32:   _int32,
		reflect.Int64:   _int64,
		reflect.Int8:    _int8,
		reflect.Int:     _int32,
		reflect.Map:     _nvlist,
		reflect.String:  _string,
		reflect.Struct:  _nvlist,
		reflect.Uint16:  _uint16,
		reflect.Uint32:  _uint32,
		reflect.Uint64:  _uint64,
		reflect.Uint8:   _uint8,
		reflect.Uint:    _uint32,
	}

	sliceTypes = map[reflect.Kind]dataType{
		reflect.Bool:   _booleanArray,
		reflect.Int16:  _int16Array,
		reflect.Int32:  _int32Array,
		reflect.Int64:  _int64Array,
		reflect.Int8:   _int8Array,
		reflect.Int:    _int32Array,
		reflect.Map:    _nvlistArray,
		reflect.String: _stringArray,
		reflect.Struct: _nvlistArray,
		reflect.Uint16: _uint16Array,
		reflect.Uint32: _uint32Array,
		reflect.Uint64: _uint64Array,
		reflect.Uint8:  _uint8Array,
		reflect.Uint:   _uint32Array,
	}
)

func validValue(v reflect.Value) error {
	if !v.IsValid() {
		return errors.Newv("invalid type", map[string]interface{}{"type": v.Kind().String()})
	}

	return nil
}

func encodePreamble(w io.Writer, codec codec, order endianness) error {
	enc := encoding{Encoding: codec, Endianess: order}
	err := errors.Wrap(binary.Write(w, binary.BigEndian, enc), "failed to encode preamble")
	return errors.Wrapv(err, map[string]interface{}{"encoding": enc})
}

func encodeList(enc encoder, v reflect.Value) error {
	if err := enc.header(header{Flag: uniqueName}); err != nil {
		return err
	}

	v = deref(v)
	switch v.Kind() {
	case reflect.Struct:
		if err := encodeStruct(enc, v); err != nil {
			return err
		}
	case reflect.Map:
		keys := make([]string, len(v.MapKeys()))
		for i, k := range v.MapKeys() {
			keys[i] = k.Interface().(string)
		}
		sort.Strings(keys)

		for _, name := range keys {
			val := v.MapIndex(reflect.ValueOf(name))
			if err := encodeItem(enc, name, nil, val); err != nil {
				return err
			}
		}
	default:
		return errors.Newv("invalid type, must be a struct", map[string]interface{}{"type": v.Kind().String()})
	}

	return enc.footer()
}

func encodeStruct(enc encoder, v reflect.Value) error {
	var err error

	forEachField(v, func(i int, field reflect.Value) bool {
		// Skip fields that can't be set (e.g. unexported)
		if !field.CanSet() {
			return true
		}
		name := v.Type().Field(i).Name
		tags := getTags(i, v)
		// Skip fields tagged as "extra"
		if len(tags) > 1 && tags[1] == "extra" {
			return true
		}
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}

		if err = encodeItem(enc, name, tags, field); err != nil {
			return false
		}
		return true
	})

	return err
}

func encodeItem(enc encoder, name string, tags []string, field reflect.Value) error {
	var tagType dataType
	if len(tags) > 1 {
		if tags[1] == "byte" {
			tagType = _byte
		} else if tags[1] == "uint8" {
			tagType = _uint8
		}
	}

	field = deref(field)
	dtype, ok := types[field.Kind()]

	switch field.Kind() {
	case reflect.Bool:
		if field.Type().Name() == "Boolean" {
			dtype = _boolean
		}
	case reflect.Interface:
		return encodeItem(enc, name, tags, reflect.ValueOf(field.Interface()))
	case reflect.Slice, reflect.Array:
		dtype, ok = sliceTypes[field.Type().Elem().Kind()]
		switch tagType {
		case _byte:
			dtype = _byteArray
		case _uint8:
			dtype = _uint8Array
		}
	case reflect.Int64:
		if field.Type().String() == "time.Duration" {
			dtype = _hrtime
		}
	case reflect.Uint8:
		switch tagType {
		case _byte:
			dtype = _byte
		case _uint8:
			dtype = _uint8
		}
	}

	if !ok {
		return errors.Newv("unknown type", map[string]interface{}{"type": field.Kind()})
	}

	return enc.item(name, dtype, field.Interface())
}
