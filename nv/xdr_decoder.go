package nv

import (
	"io"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

type xdrDecoder struct {
	*xdr.Decoder
}

func newXDRDecoder(r io.Reader) xdrDecoder {
	return xdrDecoder{Decoder: xdr.NewDecoder(r)}
}

func (d xdrDecoder) DecodeBool() (bool, error) {
	i := false
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeByte() (byte, error) {
	i := int(0)
	_, err := d.Decode(&i)
	return byte(i & 0xff), err
}

func (d xdrDecoder) DecodeInt8() (int8, error) {
	i := int8(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeInt16() (int16, error) {
	i := int16(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeInt32() (int32, error) {
	i := int32(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeInt64() (int64, error) {
	i := int64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeUint8() (uint8, error) {
	i, _, err := d.DecodeInt()
	return uint8(i & 0xff), err
}

func (d xdrDecoder) DecodeUint16() (uint16, error) {
	i := uint16(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeUint32() (uint32, error) {
	i := uint32(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeUint64() (uint64, error) {
	i := uint64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeHRTime() (time.Duration, error) {
	i := uint64(0)
	_, err := d.Decode(&i)
	return time.Duration(i), err
}

func (d xdrDecoder) DecodeString() (string, error) {
	i := ""
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeFloat64() (float64, error) {
	i := float64(0)
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeBoolArray() ([]bool, error) {
	i := []bool{}
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeByteArray() ([]byte, error) {
	i := []byte{}
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeInt8Array() ([]int8, error) {
	t := []int32{}
	_, err := d.Decode(&t)
	out := make([]int8, len(t))
	for i := range t {
		out[i] = int8(t[i])
	}
	return out, err
}

func (d xdrDecoder) DecodeInt16Array() ([]int16, error) {
	t := []int32{}
	_, err := d.Decode(&t)
	out := make([]int16, len(t))
	for i := range t {
		out[i] = int16(t[i])
	}
	return out, err
}

func (d xdrDecoder) DecodeInt32Array() ([]int32, error) {
	out := []int32{}
	_, err := d.Decode(&out)
	return out, err
}

func (d xdrDecoder) DecodeInt64Array() ([]int64, error) {
	i := []int64{}
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeUint8Array() ([]uint8, error) {
	t := []uint32{}
	_, err := d.Decode(&t)
	out := make([]uint8, len(t))
	for i := range t {
		out[i] = uint8(t[i])
	}
	return out, err
}

func (d xdrDecoder) DecodeUint16Array() ([]uint16, error) {
	t := []uint32{}
	_, err := d.Decode(&t)
	out := make([]uint16, len(t))
	for i := range t {
		out[i] = uint16(t[i])
	}
	return out, err
}

func (d xdrDecoder) DecodeUint32Array() ([]uint32, error) {
	i := []uint32{}
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeUint64Array() ([]uint64, error) {
	i := []uint64{}
	_, err := d.Decode(&i)
	return i, err
}

func (d xdrDecoder) DecodeStringArray() ([]string, error) {
	i := []string{}
	_, err := d.Decode(&i)
	return i, err
}
