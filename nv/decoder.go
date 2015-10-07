package nv

import (
	"io"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

type decoder struct {
	*xdr.Decoder
}

func newDecoder(r io.Reader) decoder {
	return decoder{Decoder: xdr.NewDecoder(r)}
}

func (d decoder) DecodeBool() (bool, error) {
	i := false
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeByte() (byte, error) {
	i := int(0)
	_, err := d.Decode(&i)
	return byte(i & 0xff), err
}

func (d decoder) DecodeInt8() (int8, error) {
	i := int8(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeInt16() (int16, error) {
	i := int16(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeInt32() (int32, error) {
	i := int32(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeInt64() (int64, error) {
	i := int64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeUint8() (uint8, error) {
	i, _, err := d.DecodeInt()
	return uint8(i & 0xff), err
}

func (d decoder) DecodeUint16() (uint16, error) {
	i := uint16(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeUint32() (uint32, error) {
	i := uint32(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeUint64() (uint64, error) {
	i := uint64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeHRTime() (time.Duration, error) {
	i := uint64(0)
	_, err := d.Decode(&i)
	return time.Duration(i), err
}

func (d decoder) DecodeString() (string, error) {
	i := ""
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeFloat64() (float64, error) {
	i := float64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeBoolArray() ([]bool, error) {
	i := []bool{}
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeByteArray() ([]byte, error) {
	i := []byte{}
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeInt8Array() ([]int8, error) {
	t := []int32{}
	_, err := d.Decode(&t)
	out := make([]int8, len(t))
	for i := range t {
		out[i] = int8(t[i])
	}
	return out, err
}

func (d decoder) DecodeInt16Array() ([]int16, error) {
	t := []int32{}
	_, err := d.Decode(&t)
	out := make([]int16, len(t))
	for i := range t {
		out[i] = int16(t[i])
	}
	return out, err
}

func (d decoder) DecodeInt32Array() ([]int32, error) {
	out := []int32{}
	_, err := d.Decode(&out)
	return out, err
}

func (d decoder) DecodeInt64Array() ([]int64, error) {
	i := []int64{}
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeUint8Array() ([]uint8, error) {
	t := []uint32{}
	_, err := d.Decode(&t)
	out := make([]uint8, len(t))
	for i := range t {
		out[i] = uint8(t[i])
	}
	return out, err
}

func (d decoder) DecodeUint16Array() ([]uint16, error) {
	t := []uint32{}
	_, err := d.Decode(&t)
	out := make([]uint16, len(t))
	for i := range t {
		out[i] = uint16(t[i])
	}
	return out, err
}

func (d decoder) DecodeUint32Array() ([]uint32, error) {
	i := []uint32{}
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeUint64Array() ([]uint64, error) {
	i := []uint64{}
	_, err := d.Decode(&i)
	return i, err
}

func (d decoder) DecodeStringArray() ([]string, error) {
	i := []string{}
	_, err := d.Decode(&i)
	return i, err
}
