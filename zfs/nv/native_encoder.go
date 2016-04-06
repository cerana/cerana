package nv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

type nativeEncoder struct {
	w          io.Writer
	order      binary.ByteOrder
	embedded   bool
	skipHeader bool
}

func NewNativeEncoder(w io.Writer) *nativeEncoder {
	return &nativeEncoder{w: w, order: binary.LittleEndian}
}

func (e nativeEncoder) Encode(i interface{}) error {
	if i == nil {
		return errors.New("can not encode a nil pointer")
	}

	v := reflect.ValueOf(i)

	if err := validValue(v); err != nil {
		return err
	}

	if err := encodePreamble(e.w, nativeCodec, littleEndian); err != nil {
		return err
	}

	return encodeList(e, v)
}

func (e nativeEncoder) header(h header) error {
	if e.skipHeader {
		return nil
	}
	if err := binary.Write(e.w, binary.LittleEndian, h); err != nil {
		return err
	}
	if e.embedded {
		if _, err := e.w.Write(empty8[:]); err != nil {
			return err
		}
		if _, err := e.w.Write(empty8[:]); err != nil {
			return err
		}
	}
	return nil
}

func (e nativeEncoder) footer() error {
	return binary.Write(e.w, e.order, uint32(0))
}

func (e nativeEncoder) item(name string, dtype dataType, value interface{}) error {
	/*
		type nativePair struct {
			Size      uint32
			NameLen   uint32
			NElements uint32
			Type      dataType
			Name      []byte
			Data      interface{}
		}
	*/
	nelements := uint32(1)

	//data := value
	vbuf := &bytes.Buffer{}
	size := 0
	switch dtype {
	case _boolean:
		nelements = 0
	case _booleanValue:
		size = 4
		if value.(bool) {
			value = uint32(1)
		} else {
			value = uint32(0)
		}
	case _booleanArray:
		nelements = uint32(len(value.([]bool)))
		size = int(4 * nelements)
		bools := value.([]bool)
		uint32s := make([]uint32, nelements)
		for i := uint32(0); i < nelements; i++ {
			if bools[i] {
				uint32s[i] = 1
			} else {
				uint32s[i] = 0
			}
		}
		value = uint32s
	case _int8, _uint8:
		size = 1
	case _int16, _uint16:
		size = 2
	case _byte:
		value = int32(value.(byte))
		fallthrough
	case _int32, _uint32:
		size = 4
	case _int64, _uint64, _double, _hrtime:
		size = 8
	case _int8Array:
		nelements = uint32(len(value.([]int8)))
		size = int(1 * nelements)
	case _int16Array:
		nelements = uint32(len(value.([]int16)))
		size = int(2 * nelements)
	case _int32Array:
		nelements = uint32(len(value.([]int32)))
		size = int(4 * nelements)
	case _int64Array:
		nelements = uint32(len(value.([]int64)))
		size = int(8 * nelements)
	case _byteArray, _uint8Array:
		nelements = uint32(len(value.([]uint8)))
		size = int(1 * nelements)
	case _uint16Array:
		nelements = uint32(len(value.([]uint16)))
		size = int(2 * nelements)
	case _uint32Array:
		nelements = uint32(len(value.([]uint32)))
		size = int(4 * nelements)
	case _uint64Array:
		nelements = uint32(len(value.([]uint64)))
		size = int(8 * nelements)
	case _string:
		str := value.(string)
		size = len(str) + 1
		buf := make([]byte, size)
		copy(buf, str)
		buf = buf[:size]
		value = buf
	case _stringArray:
		nelements = uint32(len(value.([]string)))
		strs := value.([]string)
		size = int(nelements * 8)
		i := size
		for i := range strs {
			strlen := len(strs[i]) + 1
			size += strlen
		}
		buf := make([]byte, size)
		for _, str := range strs {
			copy(buf[i:], str)
			i += len(str) + 1
		}
		value = buf
	case _nvlist:
		enc := NewNativeEncoder(vbuf)
		enc.embedded = true
		if err := encodeList(enc, reflect.ValueOf(value)); err != nil {
			return err
		}
		// embedded's nvl header + double interior pointers
		size = 24
	case _nvlistArray:
		nelements = uint32(len(value.([]map[string]interface{})))
		// 24 for the nvlist + 8 bytes for the pointer as part of the
		// array
		size = int(32 * nelements)
		for i := uint32(0); i < nelements; i++ {
			if _, err := vbuf.Write(empty8[:]); err != nil {
				return err
			}
		}
		enc := NewNativeEncoder(vbuf)
		enc.embedded = true
		for i := uint32(0); i < nelements; i++ {
			if err := enc.header(header{Flag: uniqueName}); err != nil {
				return err
			}
		}
		enc.skipHeader = true
		for _, l := range value.([]map[string]interface{}) {
			if err := encodeList(enc, reflect.ValueOf(l)); err != nil {
				return err
			}
		}
		value = []byte{}
	}

	sizeAligned := uint32(align8(size))
	if vbuf.Len() == 0 && dtype != _boolean {
		if err := binary.Write(vbuf, e.order, value); err != nil {
			return err
		}
		diff := sizeAligned - uint32(size)
		if _, err := vbuf.Write(empty8[:diff]); err != nil {
			return err
		}
	}

	pbuf := &bytes.Buffer{}
	nameLen := uint32(len(name) + 1)
	nameLenAligned := uint32(align8(int(nameLen)))
	dataSize := nameLenAligned + 16 + sizeAligned

	if err := binary.Write(pbuf, e.order, dataSize); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if err := binary.Write(pbuf, e.order, nameLen); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if err := binary.Write(pbuf, e.order, nelements); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if err := binary.Write(pbuf, e.order, dtype); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if err := binary.Write(pbuf, e.order, []byte(name)); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if _, err := pbuf.Write(empty8[:nameLenAligned-nameLen+1]); err != nil {
		return err
	} else if _, err = pbuf.WriteTo(e.w); err != nil {
		return err
	}

	if _, err := vbuf.WriteTo(e.w); err != nil {
		return err
	}

	return nil
}
