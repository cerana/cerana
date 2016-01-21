package nv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

type XDREncoder struct {
	w io.Writer
}

func NewXDREncoder(w io.Writer) XDREncoder {
	return XDREncoder{w: w}
}

func (e XDREncoder) Encode(i interface{}) error {
	if i == nil {
		return errors.New("can not encode a nil pointer")
	}

	v := reflect.ValueOf(i)

	if err := validValue(v); err != nil {
		return err
	}

	if err := encodePreamble(e.w, xdrCodec, littleEndian); err != nil {
		return err
	}

	return encodeList(e, v)
}

func (e XDREncoder) header(h header) error {
	return encHeader(e.w, h)
}

func encHeader(w io.Writer, h header) error {
	return binary.Write(w, binary.BigEndian, h)
}

func (e XDREncoder) footer() error {
	return encFooter(e.w)
}

func encFooter(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, uint64(0))
}

func (e XDREncoder) item(name string, dtype dataType, value interface{}) error {
	p := pair{
		Name:      name,
		NElements: 1,
		Type:      dtype,
		data:      value,
	}

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
		enc := NewXDREncoder(vbuf)
		if err := encodeList(enc, reflect.ValueOf(value)); err != nil {
			return err
		}
		p.data = vbuf.Bytes()
	case _NVLIST_ARRAY:
		p.NElements = uint32(len(value.([]map[string]interface{})))
		for _, l := range value.([]map[string]interface{}) {
			enc := NewXDREncoder(vbuf)
			if err := encodeList(enc, reflect.ValueOf(l)); err != nil {
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

	_, err = pbuf.WriteTo(e.w)
	if err != nil {
		return err
	}
	_, err = vbuf.WriteTo(e.w)
	if err != nil {
		return err
	}

	return nil
}
