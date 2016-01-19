package nv

import "time"

func (d xdrDecoder) decodeBool() (bool, error) {
	i := false
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeByte() (byte, error) {
	i := int(0)
	_, err := d.Decoder.Decode(&i)
	return byte(i & 0xff), err
}

func (d xdrDecoder) decodeInt8() (int8, error) {
	i := int8(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeInt16() (int16, error) {
	i := int16(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeInt32() (int32, error) {
	i := int32(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeInt64() (int64, error) {
	i := int64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeUint8() (uint8, error) {
	i, _, err := d.Decoder.DecodeInt()
	return uint8(i & 0xff), err
}

func (d xdrDecoder) decodeUint16() (uint16, error) {
	i := uint16(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeUint32() (uint32, error) {
	i := uint32(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeUint64() (uint64, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeHRTime() (time.Duration, error) {
	i := uint64(0)
	_, err := d.Decoder.Decode(&i)
	return time.Duration(i), err
}

func (d xdrDecoder) decodeString() (string, error) {
	i := ""
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeFloat64() (float64, error) {
	i := float64(0)
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeBoolArray() ([]bool, error) {
	i := []bool{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeByteArray() ([]byte, error) {
	i := []byte{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeInt8Array() ([]int8, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int8, len(t))
	for i := range t {
		out[i] = int8(t[i])
	}
	return out, err
}

func (d xdrDecoder) decodeInt16Array() ([]int16, error) {
	t := []int32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]int16, len(t))
	for i := range t {
		out[i] = int16(t[i])
	}
	return out, err
}

func (d xdrDecoder) decodeInt32Array() ([]int32, error) {
	out := []int32{}
	_, err := d.Decoder.Decode(&out)
	return out, err
}

func (d xdrDecoder) decodeInt64Array() ([]int64, error) {
	i := []int64{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeUint8Array() ([]uint8, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint8, len(t))
	for i := range t {
		out[i] = uint8(t[i])
	}
	return out, err
}

func (d xdrDecoder) decodeUint16Array() ([]uint16, error) {
	t := []uint32{}
	_, err := d.Decoder.Decode(&t)
	out := make([]uint16, len(t))
	for i := range t {
		out[i] = uint16(t[i])
	}
	return out, err
}

func (d xdrDecoder) decodeUint32Array() ([]uint32, error) {
	i := []uint32{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeUint64Array() ([]uint64, error) {
	i := []uint64{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}

func (d xdrDecoder) decodeStringArray() ([]string, error) {
	i := []string{}
	_, err := d.Decoder.Decode(&i)
	return i, err
}
