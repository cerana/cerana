package nv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

func Decode(buf []byte) (mList, error) {
	b := bytes.NewReader(buf)

	enc := encoding{}
	err := binary.Read(b, binary.BigEndian, &enc)
	if err != nil {
		return nil, err
	}

	if enc.Encoding > 1 {
		return nil, fmt.Errorf("invalid encoding: %v", enc.Encoding)
	}
	if enc.Endianess > 1 {
		return nil, fmt.Errorf("invalid endianess: %v", enc.Endianess)
	}
	if enc.Reserved1 != 0 {
		return nil, fmt.Errorf("unexpected reserved1 value: %v", enc.Reserved1)
	}
	if enc.Reserved2 != 0 {
		return nil, fmt.Errorf("unexpected reserved2 value: %v", enc.Reserved2)
	}

	return decodeList(b)
}

func decodeList(r io.ReadSeeker) (mList, error) {
	var h header
	err := binary.Read(r, binary.BigEndian, &h)
	if err != nil {
		return nil, err
	}

	if h.Version != 0 {
		return nil, fmt.Errorf("unexpected version: %v", h.Version)
	}
	if h.Flag < UNIQUE_NAME || h.Flag > UNIQUE_NAME_TYPE {
		return nil, fmt.Errorf("unexpected Flag: %v", h.Flag)
	}

	m := mList{}
	for {
		end, err := isEnd(r)
		if err != nil {
			return nil, err
		}
		if end {
			break
		}

		p := pair{}
		_, err = xdr.Unmarshal(r, &p)
		if err != nil {
			return nil, err
		}

		var v interface{}
		dec := newDecoder(r)
		switch p.Type {
		case BOOLEAN_VALUE:
			v, err = dec.DecodeBool()
		case BYTE:
			v, err = dec.DecodeByte()
		case INT8:
			v, err = dec.DecodeInt8()
		case INT16:
			v, err = dec.DecodeInt16()
		case INT32:
			v, err = dec.DecodeInt32()
		case INT64:
			v, err = dec.DecodeInt64()
		case UINT8:
			v, err = dec.DecodeUint8()
		case UINT16:
			v, err = dec.DecodeUint16()
		case UINT32:
			v, err = dec.DecodeUint32()
		case UINT64:
			v, err = dec.DecodeUint64()
		case HRTIME:
			v, err = dec.DecodeHRTime()
		case DOUBLE:
			v, err = dec.DecodeFloat64()
		case BOOLEAN_ARRAY:
			v, err = dec.DecodeBoolArray()
		case BYTE_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.DecodeByteArray()
			}
		case INT8_ARRAY:
			v, err = dec.DecodeInt8Array()
		case INT16_ARRAY:
			v, err = dec.DecodeInt16Array()
		case INT32_ARRAY:
			v, err = dec.DecodeInt32Array()
		case INT64_ARRAY:
			v, err = dec.DecodeInt64Array()
		case UINT8_ARRAY:
			v, err = dec.DecodeUint8Array()
		case UINT16_ARRAY:
			v, err = dec.DecodeUint16Array()
		case UINT32_ARRAY:
			v, err = dec.DecodeUint32Array()
		case UINT64_ARRAY:
			v, err = dec.DecodeUint64Array()
		case STRING:
			v, err = dec.DecodeString()
		case STRING_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.DecodeStringArray()
			}
		case NVLIST:
			v, err = decodeList(r)
		case NVLIST_ARRAY:
			arr := make([]mList, 0, p.NElements)
			for i := uint32(0); i < p.NElements; i++ {
				var m mList
				m, err = decodeList(r)
				if err != nil {
					break
				}
				arr = append(arr, m)
			}
			v = arr
		default:
			return nil, fmt.Errorf("unknown type: %v", p.Type)
		}
		if err != nil {
			return nil, err
		}

		m[p.Name] = mVal{
			Name:  p.Name,
			Type:  p.Type,
			Value: v,
		}

	}
	return m, nil
}
