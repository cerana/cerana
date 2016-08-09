package nv

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"

	"github.com/cerana/cerana/pkg/errors"
	xdr "github.com/davecgh/go-xdr/xdr2"
)

// XDREncoder is an Encoder for XDR encoding.
type XDREncoder struct {
	w io.Writer
}

// NewXDREncoder creates a new XDREncoder.
func NewXDREncoder(w io.Writer) XDREncoder {
	return XDREncoder{w: w}
}

// Encode encodes data using XDR encoding.
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
	return errors.Wrapv(binary.Write(w, binary.BigEndian, h), map[string]interface{}{"header": h})
}

func (e XDREncoder) footer() error {
	return encFooter(e.w)
}

func encFooter(w io.Writer) error {
	return errors.Wrap(binary.Write(w, binary.BigEndian, uint64(0)))
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
	case _boolean:
		p.NElements = 0
	case _byte:
		value = int8(value.(uint8))
	case _uint8:
		value = int(int8(value.(uint8)))
	case _byteArray:
		p.NElements = uint32(len(value.([]byte)))
		n := int(p.NElements)
		arrType := reflect.ArrayOf(n, reflect.TypeOf(byte(0)))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]byte) {
			arr.Index(i).SetUint(uint64(b))
		}
		value = arr.Interface()
	case _booleanArray:
		p.NElements = uint32(len(value.([]bool)))
	case _int8Array:
		p.NElements = uint32(len(value.([]int8)))
	case _int16Array:
		p.NElements = uint32(len(value.([]int16)))
	case _int32Array:
		p.NElements = uint32(len(value.([]int32)))
	case _int64Array:
		p.NElements = uint32(len(value.([]int64)))
	case _uint8Array:
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
	case _uint16Array:
		p.NElements = uint32(len(value.([]uint16)))
	case _uint32Array:
		p.NElements = uint32(len(value.([]uint32)))
	case _uint64Array:
		p.NElements = uint32(len(value.([]uint64)))
	case _stringArray:
		p.NElements = uint32(len(value.([]string)))
		arrType := reflect.ArrayOf(int(p.NElements), reflect.TypeOf(""))
		arr := reflect.New(arrType).Elem()
		for i, b := range value.([]string) {
			arr.Index(i).SetString(b)
		}
		value = arr.Interface()
	case _nvlist:
		enc := NewXDREncoder(vbuf)
		if err := encodeList(enc, reflect.ValueOf(value)); err != nil {
			return err
		}
		p.data = vbuf.Bytes()
	case _nvlistArray:
		p.NElements = uint32(len(value.([]map[string]interface{})))
		for _, l := range value.([]map[string]interface{}) {
			enc := NewXDREncoder(vbuf)
			if err := encodeList(enc, reflect.ValueOf(l)); err != nil {
				return err
			}
		}
		p.data = vbuf.Bytes()
	}

	if vbuf.Len() == 0 && p.Type != _boolean {
		_, err := xdr.NewEncoder(vbuf).Encode(value)
		if err != nil {
			return errors.Wrap(err, "failed to write value to vbuf")
		}
	}

	p.EncodedSize = uint32(p.encodedSize())
	p.DecodedSize = uint32(p.decodedSize())

	pbuf := &bytes.Buffer{}
	_, err := xdr.NewEncoder(pbuf).Encode(p)
	if err != nil {
		return errors.Wrap(err, "failed to write to pbuf")
	}

	_, err = pbuf.WriteTo(e.w)
	if err != nil {
		return errors.Wrap(err, "failed to write pbuf")
	}
	_, err = vbuf.WriteTo(e.w)
	if err != nil {
		return errors.Wrap(err, "failed to write vbuf")
	}

	return nil
}
