package nv

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/cerana/cerana/pkg/errors"
)

func (d NativeDecoder) decode(i interface{}) error {
	return errors.Wrap(binary.Read(d.r, binary.LittleEndian, i))
}

func (d NativeDecoder) decodeBool() (bool, error) {
	i := uint32(0)
	err := d.decode(&i)
	return i != 0, err
}

func (d NativeDecoder) decodeByte() (byte, error) {
	i := uint8(0)
	err := d.decode(&i)
	return byte(i & 0xff), err
}

func (d NativeDecoder) decodeInt() (int, error) {
	i := int32(0)
	err := d.decode(&i)
	return int(i), err
}

func (d NativeDecoder) decodeInt8() (int8, error) {
	i := int8(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt16() (int16, error) {
	i := int16(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt32() (int32, error) {
	i := int32(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt64() (int64, error) {
	i := int64(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint8() (uint8, error) {
	i, err := d.decodeInt()
	return uint8(i & 0xff), err
}

func (d NativeDecoder) decodeUint16() (uint16, error) {
	i := uint16(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint32() (uint32, error) {
	i := uint32(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint64() (uint64, error) {
	i := uint64(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeHRTime() (time.Duration, error) {
	i := uint64(0)
	err := d.decode(&i)
	return time.Duration(i), err
}

func (d NativeDecoder) decodeString(size int) (string, error) {
	i := make([]byte, size)
	err := d.decode(&i)
	index := bytes.IndexByte(i, byte(0))
	s := string(i[:index])
	return s, err
}

func (d NativeDecoder) decodeFloat64() (float64, error) {
	i := float64(0)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeBoolArray(nelem uint32) ([]bool, error) {
	uints := make([]uint32, nelem)
	err := d.decode(&uints)
	if err != nil {
		return nil, err
	}

	bools := make([]bool, nelem)
	for i := range uints {
		bools[i] = uints[i] != 0
	}
	return bools, err
}

func (d NativeDecoder) decodeByteArray(nelem uint32) ([]byte, error) {
	i := make([]byte, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt8Array(nelem uint32) ([]int8, error) {
	i := make([]int8, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt16Array(nelem uint32) ([]int16, error) {
	i := make([]int16, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt32Array(nelem uint32) ([]int32, error) {
	i := make([]int32, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeInt64Array(nelem uint32) ([]int64, error) {
	i := make([]int64, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint8Array(nelem uint32) ([]uint8, error) {
	i := make([]uint8, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint16Array(nelem uint32) ([]uint16, error) {
	i := make([]uint16, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint32Array(nelem uint32) ([]uint32, error) {
	i := make([]uint32, nelem)
	err := d.decode(&i)
	return i, err
}

func (d NativeDecoder) decodeUint64Array(nelem uint32) ([]uint64, error) {
	i := make([]uint64, nelem)
	err := d.decode(&i)
	return i, err
}

// adapted from bufio ScanLines
func splitCString(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, 0); i > 0 {
		// We have a full newline-terminated line.
		return i + 1, data[:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), nil, io.EOF
	}
	// Request more data.
	return 0, nil, nil
}

func (d NativeDecoder) decodeStringArray(len int64, nelem uint32) ([]string, error) {
	i := make([]string, 0, nelem)
	r := &io.LimitedReader{R: d.r, N: len}
	scanner := bufio.NewScanner(r)
	scanner.Split(splitCString)
	for scanner.Scan() {
		b := scanner.Bytes()
		s := string(b)
		i = append(i, s)
	}
	return i, errors.Wrap(scanner.Err())
}
