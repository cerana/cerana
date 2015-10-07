package nv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

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

		for _, k := range keys {
			_, err = encodeItem(w, k, v.MapIndex(reflect.ValueOf(k)))
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
	numFields := v.NumField()
	for i := 0; i < numFields; i++ {
		field := v.Field(i)

		structField := v.Type().Field(i)
		name := structField.Name
		if tag := structField.Tag.Get("nv"); tag != "" {
			tags := strings.Split(tag, ",")
			if len(tags) > 0 && tags[0] != "" {
				name = tags[0]
			}
		}

		encodeItem(w, name, field)
	}

	if err = binary.Write(w, binary.BigEndian, uint64(0)); err != nil {
		return 0, err
	}
	return size + 8, nil
}

func encodeItem(w io.Writer, name string, field reflect.Value) ([]byte, error) {
	p := pair{
		Name:      name,
		NElements: 1,
	}

	switch t := field.Kind(); t {
	case reflect.String:
		p.Type = STRING
	case reflect.Uint64:
		p.Type = UINT64
	case reflect.Int32, reflect.Int:
		p.Type = INT32
	case reflect.Struct:
		p.Type = NVLIST
	default:
		return nil, fmt.Errorf("unknown type:", t)
	}

	if field.Type().String() == "nv.mVal" {
		p.Type = field.FieldByName("Type").Interface().(dataType)
		p.Name = field.FieldByName("Name").Interface().(string)
		p.data = field.FieldByName("Value").Interface()
	} else {
		if !field.CanSet() {
			panic("can't Set")
			return nil, nil
		}
		p.data = field.Interface()
	}

	if p.Type == UNKNOWN || p.Type > DOUBLE {
		return nil, fmt.Errorf("invalid Type '%v'", p.Type)
	}

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
		p.NElements = uint32(len(value.([]mList)))
		for _, l := range value.([]mList) {
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
