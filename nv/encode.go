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
	if err = binary.Write(w, binary.BigEndian, header{Flag: UNIQUE_NAME}); err != nil {
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
			_, err = encodeItem(w, name, nil, v.MapIndex(reflect.ValueOf(name)))
			if err != nil {
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

	forEachField(v, func(i int, field reflect.Value) {
		// Skip fields that can't be set (e.g. unexported)
		if !field.CanSet() {
			return
		}
		name := v.Type().Field(i).Name
		tags := getTags(i, v)
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}
		encodeItem(w, name, tags, field)
	})

	if err = binary.Write(w, binary.BigEndian, uint64(0)); err != nil {
		return 0, err
	}
	return size + 8, nil
}

func encodeItem(w io.Writer, name string, tags []string, field reflect.Value) ([]byte, error) {
	var types = map[reflect.Kind]dataType{
		reflect.Bool:    BOOLEAN_VALUE,
		reflect.Float32: DOUBLE,
		reflect.Float64: DOUBLE,
		reflect.Int16:   INT16,
		reflect.Int32:   INT32,
		reflect.Int64:   INT64,
		reflect.Int8:    INT8,
		reflect.Int:     INT32,
		reflect.Map:     NVLIST,
		reflect.String:  STRING,
		reflect.Struct:  NVLIST,
		reflect.Uint16:  UINT16,
		reflect.Uint32:  UINT32,
		reflect.Uint64:  UINT64,
		reflect.Uint8:   UINT8,
		reflect.Uint:    UINT32,
	}

	var sliceTypes = map[reflect.Kind]dataType{
		reflect.Bool:   BOOLEAN_ARRAY,
		reflect.Int16:  INT16_ARRAY,
		reflect.Int32:  INT32_ARRAY,
		reflect.Int64:  INT64_ARRAY,
		reflect.Int8:   INT8_ARRAY,
		reflect.Int:    INT32_ARRAY,
		reflect.Map:    NVLIST_ARRAY,
		reflect.String: STRING_ARRAY,
		reflect.Struct: NVLIST_ARRAY,
		reflect.Uint16: UINT16_ARRAY,
		reflect.Uint32: UINT32_ARRAY,
		reflect.Uint64: UINT64_ARRAY,
		reflect.Uint8:  UINT8_ARRAY,
		reflect.Uint:   UINT32_ARRAY,
	}
	var tagType dataType
	if len(tags) > 1 {
		if tags[1] == "byte" {
			tagType = BYTE
		} else if tags[1] == "uint8" {
			tagType = UINT8
		}
	}

	p := pair{
		Name:      name,
		NElements: 1,
	}

	var ok bool
	p.Type, ok = types[field.Kind()]

	switch field.Kind() {
	case reflect.Interface:
		return encodeItem(w, name, tags, reflect.ValueOf(field.Interface()))
	case reflect.Slice, reflect.Array:
		p.Type, ok = sliceTypes[field.Index(0).Kind()]
		switch tagType {
		case BYTE:
			p.Type = BYTE_ARRAY
		case UINT8:
			p.Type = UINT8_ARRAY
		}
	case reflect.Int64:
		if field.Type().String() == "time.Duration" {
			p.Type = HRTIME
		}
	case reflect.Uint8:
		switch tagType {
		case BYTE:
			p.Type = BYTE
		case UINT8:
			p.Type = UINT8
		}
	}

	if !ok {
		return nil, fmt.Errorf("unknown type: %v", field.Kind())
	}

	p.data = field.Interface()
	value := p.data
	vbuf := &bytes.Buffer{}
	switch p.Type {
	case BYTE:
		value = int8(value.(uint8))
	case UINT8:
		value = int(int8(value.(uint8)))
	case BYTE_ARRAY:
		p.NElements = uint32(len(value.([]byte)))
		n := int(p.NElements)
		arrType := reflect.ArrayOf(n, reflect.TypeOf(byte(0)))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]byte) {
			arr.Index(i).SetUint(uint64(b))
		}
		value = arr.Interface()
	case BOOLEAN_ARRAY:
		p.NElements = uint32(len(value.([]bool)))
	case INT8_ARRAY:
		p.NElements = uint32(len(value.([]int8)))
	case INT16_ARRAY:
		p.NElements = uint32(len(value.([]int16)))
	case INT32_ARRAY:
		p.NElements = uint32(len(value.([]int32)))
	case INT64_ARRAY:
		p.NElements = uint32(len(value.([]int64)))
	case UINT8_ARRAY:
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
	case UINT16_ARRAY:
		p.NElements = uint32(len(value.([]uint16)))
	case UINT32_ARRAY:
		p.NElements = uint32(len(value.([]uint32)))
	case UINT64_ARRAY:
		p.NElements = uint32(len(value.([]uint64)))
	case STRING_ARRAY:
		p.NElements = uint32(len(value.([]string)))
		arrType := reflect.ArrayOf(int(p.NElements), reflect.TypeOf(""))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]string) {
			arr.Index(i).SetString(b)
		}
		value = arr.Interface()
	case NVLIST:
		if err := encodeList(vbuf, reflect.ValueOf(value)); err != nil {
			return nil, err
		}
		p.data = vbuf.Bytes()
	case NVLIST_ARRAY:
		p.NElements = uint32(len(value.([]map[string]interface{})))
		for _, l := range value.([]map[string]interface{}) {
			if err := encodeList(vbuf, reflect.ValueOf(l)); err != nil {
				return nil, err
			}
		}
		p.data = vbuf.Bytes()
	}

	if vbuf.Len() == 0 {
		_, err := xdr.NewEncoder(vbuf).Encode(value)
		if err != nil {
			return nil, err
		}
	}

	p.EncodedSize = uint32(p.encodedSize())
	p.DecodedSize = uint32(p.decodedSize())

	pbuf := &bytes.Buffer{}
	_, err := xdr.NewEncoder(pbuf).Encode(p)
	if err != nil {
		return nil, err
	}

	_, err = pbuf.WriteTo(w)
	if err != nil {
		return nil, err
	}
	_, err = vbuf.WriteTo(w)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
