package nv

import "time"

func (d XDRDecoder) decodeBool() (bool, error) {
	i := false
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeByte() (byte, error) {
	i := int(0)
	_, err := d.Decoder.Decode(&i)
	return byte(i & 0xff), err
}

func (d XDRDecoder) decodeInt8() (int8, error) {
	i := int8(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeInt16() (int16, error) {
	i := int16(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeInt32() (int32, error) {
	i := int32(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeInt64() (int64, error) {
	i := int64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeUint8() (uint8, error) {
	i, _, err := d.Decoder.DecodeInt()
	return uint8(i & 0xff), err
}

func (d XDRDecoder) decodeUint16() (uint16, error) {
	i := uint16(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeUint32() (uint32, error) {
	i := uint32(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeUint64() (uint64, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeHRTime() (time.Duration, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return time.Duration(i), err
}

func (d XDRDecoder) decodeString() (string, error) {
	i := ""
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeFloat64() (float64, error) {
	i := float64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeBoolArray() ([]bool, error) {
	i := []bool{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeByteArray() ([]byte, error) {
	i := []byte{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeInt8Array() ([]int8, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int8, len(t))
	for i := range t {
		out[i] = int8(t[i])
	}
	return out, err
}

func (d XDRDecoder) decodeInt16Array() ([]int16, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int16, len(t))
	for i := range t {
		out[i] = int16(t[i])
	}
	return out, err
}

func (d XDRDecoder) decodeInt32Array() ([]int32, error) {
	out := []int32{}
	_, err := d.Decoder.Decode(&out)
	return out, err
}

func (d XDRDecoder) decodeInt64Array() ([]int64, error) {
	i := []int64{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeUint8Array() ([]uint8, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint8, len(t))
	for i := range t {
		out[i] = uint8(t[i])
	}
	return out, err
}

func (d XDRDecoder) decodeUint16Array() ([]uint16, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint16, len(t))
	for i := range t {
		out[i] = uint16(t[i])
	}
	return out, err
}

func (d XDRDecoder) decodeUint32Array() ([]uint32, error) {
	i := []uint32{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeUint64Array() ([]uint64, error) {
	i := []uint64{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d XDRDecoder) decodeStringArray() ([]string, error) {
	i := []string{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}
