package nv

import (
	"time"

	"github.com/cerana/cerana/pkg/errors"
)

func (d XDRDecoder) decodeBool() (bool, error) {
	i := false
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode bool")
}

func (d XDRDecoder) decodeByte() (byte, error) {
	i := int(0)
	_, err := d.Decoder.Decode(&i)
	return byte(i & 0xff), errors.Wrap(err, "failed to decode byte")
}

func (d XDRDecoder) decodeInt8() (int8, error) {
	i := int8(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode int8")
}

func (d XDRDecoder) decodeInt16() (int16, error) {
	i := int16(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode int16")
}

func (d XDRDecoder) decodeInt32() (int32, error) {
	i := int32(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode int32")
}

func (d XDRDecoder) decodeInt64() (int64, error) {
	i := int64(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode int64")
}

func (d XDRDecoder) decodeUint8() (uint8, error) {
	i, _, err := d.Decoder.DecodeInt()
	return uint8(i & 0xff), errors.Wrap(err, "failed to decode uint8")
}

func (d XDRDecoder) decodeUint16() (uint16, error) {
	i := uint16(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode uint16")
}

func (d XDRDecoder) decodeUint32() (uint32, error) {
	i := uint32(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode uint32")
}

func (d XDRDecoder) decodeUint64() (uint64, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode uint64")
}

func (d XDRDecoder) decodeHRTime() (time.Duration, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return time.Duration(i), errors.Wrap(err, "failed to decode hrtime")
}

func (d XDRDecoder) decodeString() (string, error) {
	i := ""
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode string")
}

func (d XDRDecoder) decodeFloat64() (float64, error) {
	i := float64(0)
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode float64")
}

func (d XDRDecoder) decodeBoolArray() ([]bool, error) {
	i := []bool{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode bool array")
}

func (d XDRDecoder) decodeByteArray() ([]byte, error) {
	i := []byte{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode byte array")
}

func (d XDRDecoder) decodeInt8Array() ([]int8, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int8, len(t))
	for i := range t {
		out[i] = int8(t[i])
	}
	return out, errors.Wrap(err, "failed to decode int8 array")
}

func (d XDRDecoder) decodeInt16Array() ([]int16, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int16, len(t))
	for i := range t {
		out[i] = int16(t[i])
	}
	return out, errors.Wrap(err, "failed to decode int16 array")
}

func (d XDRDecoder) decodeInt32Array() ([]int32, error) {
	out := []int32{}
	_, err := d.Decoder.Decode(&out)
	return out, errors.Wrap(err, "failed to decode int32 array")
}

func (d XDRDecoder) decodeInt64Array() ([]int64, error) {
	i := []int64{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode int64 array")
}

func (d XDRDecoder) decodeUint8Array() ([]uint8, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint8, len(t))
	for i := range t {
		out[i] = uint8(t[i])
	}
	return out, errors.Wrap(err, "failed to decode uint8 array")
}

func (d XDRDecoder) decodeUint16Array() ([]uint16, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint16, len(t))
	for i := range t {
		out[i] = uint16(t[i])
	}
	return out, errors.Wrap(err, "failed to decode uint16 array")
}

func (d XDRDecoder) decodeUint32Array() ([]uint32, error) {
	i := []uint32{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode uint32 array")
}

func (d XDRDecoder) decodeUint64Array() ([]uint64, error) {
	i := []uint64{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode uint64 array")
}

func (d XDRDecoder) decodeStringArray() ([]string, error) {
	i := []string{}
	_, err := d.Decoder.Decode(&i)
	return i, errors.Wrap(err, "failed to decode string array")
}
