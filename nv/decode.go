package nv

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

func decodePreamble(r io.Reader, byteOrder binary.ByteOrder) (codec, endianness, error) {
	var err error
	enc := encoding{}
	if err = binary.Read(r, byteOrder, &enc); err != nil {
		return enc.Encoding, enc.Endianess, err
	}

	if enc.Encoding > maxCodec {
		err = fmt.Errorf("invalid encoding: %v", enc.Encoding)
	} else if enc.Endianess > maxEndian {
		err = fmt.Errorf("invalid endianess: %v", enc.Endianess)
	} else if enc.Reserved1 != 0 {
		err = fmt.Errorf("unexpected reserved1 value: %v", enc.Reserved1)
	} else if enc.Reserved2 != 0 {
		err = fmt.Errorf("unexpected reserved2 value: %v", enc.Reserved2)
	}
	return enc.Encoding, enc.Endianess, err
}

// Decode
// Note: care should be taken when decoding into a `map[string]interface{}` as
// bytes/uint8s (and their array forms) can not be distinguished and will be
// treated as uint8/[]uint8.
func Decode(data []byte, target interface{}) error {
	b := bytes.NewReader(data)

	// Validate data encoding
	_, _, err := decodePreamble(b, binary.BigEndian)
	if err != nil {
		return err
	}

	// Validate target
	targetV := reflect.ValueOf(target)
	if targetV.Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode into non-pointer: %v", reflect.TypeOf(targetV).String())
	}
	if targetV.IsNil() {
		return fmt.Errorf("cannot decode into nil")
	}

	return decodeList(b, reflect.Indirect(targetV))
}

// fieldSetFunc is used to set the value if the target is a field within a
// struct
type fieldSetFunc func(reflect.Value, reflect.Value)

func decHeader(r io.Reader) (header, error) {
	var h header
	err := binary.Read(r, binary.BigEndian, &h)
	return h, err
}

func decMeta(r io.Reader, meta *pair) error {
	_, err := xdr.Unmarshal(r, meta)
	return err
}

func skip(r io.ReadSeeker, meta pair) error {
	_, err := r.Seek(int64(meta.EncodedSize-uint32(meta.headerSize())), 1)
	return err
}

func decValue(r io.ReadSeeker, meta pair, tType reflect.Type) (reflect.Value, fieldSetFunc, error) {
	var val reflect.Value
	var fsf fieldSetFunc
	err := fmt.Errorf("unknown type: %v", meta.Type)

	var v interface{}
	dec := newXDRDecoder(r)
	switch meta.Type {
	case _BOOLEAN:
		err = nil
		v := Boolean(true)
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _BOOLEAN_VALUE:
		v, err = dec.decodeBool()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetBool(v.(bool))
		}
	case _BYTE:
		v, err = dec.decodeByte()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _INT8:
		v, err = dec.decodeInt8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int8)))
		}
	case _INT16:
		v, err = dec.decodeInt16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int16)))
		}
	case _INT32:
		v, err = dec.decodeInt32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(int32)))
		}
	case _INT64:
		v, err = dec.decodeInt64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(v.(int64))
		}
	case _UINT8:
		v, err = dec.decodeUint8()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint8)))
		}
	case _UINT16:
		v, err = dec.decodeUint16()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint16)))
		}
	case _UINT32:
		v, err = dec.decodeUint32()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint32)))
		}
	case _UINT64:
		v, err = dec.decodeUint64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetUint(uint64(v.(uint64)))
		}
	case _HRTIME:
		v, err = dec.decodeHRTime()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetInt(int64(v.(time.Duration)))
		}
	case _DOUBLE:
		v, err = dec.decodeFloat64()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetFloat(v.(float64))
		}
	case _BOOLEAN_ARRAY:
		v, err = dec.decodeBoolArray()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _BYTE_ARRAY:
		if _, err = r.Seek(-4, 1); err == nil {
			v, err = dec.decodeByteArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.SetBytes(v.([]byte))
			}
		}
	case _INT8_ARRAY:
		v, err = dec.decodeInt8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT16_ARRAY:
		v, err = dec.decodeInt16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT32_ARRAY:
		v, err = dec.decodeInt32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _INT64_ARRAY:
		v, err = dec.decodeInt64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT8_ARRAY:
		v, err = dec.decodeUint8Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT16_ARRAY:
		v, err = dec.decodeUint16Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT32_ARRAY:
		v, err = dec.decodeUint32Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _UINT64_ARRAY:
		v, err = dec.decodeUint64Array()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _STRING:
		v, err = dec.decodeString()
		val = reflect.ValueOf(v)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.SetString(v.(string))
		}
	case _STRING_ARRAY:
		if _, err = r.Seek(-4, 1); err == nil {
			v, err = dec.decodeStringArray()
			val = reflect.ValueOf(v)
			fsf = func(field reflect.Value, val reflect.Value) {
				field.Set(val)
			}
		}
	case _NVLIST:
		val = reflect.Indirect(reflect.New(tType))
		err = decodeList(r, val)
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	case _NVLIST_ARRAY:
		if tType.Kind() == reflect.Interface {
			tType = reflect.TypeOf([]map[string]interface{}{})
		}

		val = reflect.MakeSlice(tType, 0, int(meta.NElements))
		for i := uint32(0); i < meta.NElements; i++ {
			elem := reflect.Indirect(reflect.New(tType.Elem()))
			err = decodeList(r, elem)
			if err != nil {
				break
			}
			val = reflect.Append(val, elem)
		}
		fsf = func(field reflect.Value, val reflect.Value) {
			field.Set(val)
		}
	}
	return val, fsf, err
}

func decodeList(r io.ReadSeeker, target reflect.Value) error {
	// Validate data header
	h, err := decHeader(r)
	if err != nil {
		return err
	}
	if h.Version != 0 {
		return fmt.Errorf("unexpected version: %v", h.Version)
	}
	if h.Flag < _UNIQUE_NAME || h.Flag > _UNIQUE_NAME_TYPE {
		return fmt.Errorf("unexpected Flag: %v", h.Flag)
	}

	// Make sure the target is not a pointer. If it is, initilize and
	// dereference.
	if target.Kind() == reflect.Ptr {
		target.Set(reflect.New(target.Type().Elem()))
		target = reflect.Indirect(target)
	}

	// Maps and structs have slightly different handling
	// Interface types will be handled as maps
	isMap := (target.Kind() == reflect.Map || target.Kind() == reflect.Interface)

	if isMap {
		// Initialize the map. Can't add keys without this.
		if target.Kind() == reflect.Interface {
			target.Set(reflect.MakeMap(reflect.TypeOf(map[string]interface{}{})))
			target = target.Elem()
		} else {
			target.Set(reflect.MakeMap(target.Type()))
		}
	}

	// Special map alias that collects values and nvlist types
	isList := target.Type().String() == "nv.list"

	// For structs, create the lookup table for field name/tag -> field index
	var targetFieldIndexMap map[string]int
	if !isMap {
		targetFieldIndexMap = fieldIndexMap(target)
	}

	// Start decoding data
	for {
		// Done when there's no more data or an error has occured
		if end, err := isEnd(r); end || err != nil {
			return err
		}

		// Get next encoded data pair
		// Note: This just gets the data pair metadata. The actual value data
		// is left on the reader, to be pulled off by the decoder.
		dataPair := pair{}
		if err := decMeta(r, &dataPair); err != nil {
			return err
		}

		// If not dealing with a map, look up the corresponding struct field
		var targetField reflect.Value
		if !isMap {
			targetFieldIndex, ok := targetFieldIndexMap[dataPair.Name]
			// If there's no corresponding struct field, skip the data and move
			// on to the next data pair
			if !ok {
				if err := skip(r, dataPair); err != nil {
					return err
				}
				continue
			}
			targetField = target.Field(targetFieldIndex)
		}

		var targetType reflect.Type
		if isList {
			targetType = target.Type()
		} else if isMap {
			targetType = target.Type().Elem()
		} else {
			targetType = targetField.Type()
		}

		// value can be used directly if target is a map, if it is a
		// struct we need to use fieldSetter since the struct needs to
		// be set with a non-interface type
		value, fieldSetter, err := decValue(r, dataPair, targetType)
		if err != nil {
			return err
		}

		// Set the value appropriately
		if isMap {
			name := reflect.ValueOf(dataPair.Name)
			if isList {
				value = reflect.ValueOf(map[string]interface{}{
					"type":  dataPair.Type.String(),
					"value": value.Interface(),
				})
			}
			target.SetMapIndex(name, value)
		} else {
			fieldSetter(targetField, value)
		}
	}
	return nil
}

// fieldIndexMap creates a map of field names, with tag name overrides,
// to their index
func fieldIndexMap(v reflect.Value) map[string]int {
	vFieldIndexMap := make(map[string]int)
	forEachField(v, func(i int, field reflect.Value) {
		// Skip fields that can't be set (e.g. unexported)
		if !field.CanSet() {
			return
		}
		name := v.Type().Field(i).Name
		if tags := getTags(i, v); len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}
		vFieldIndexMap[name] = i
	})
	return vFieldIndexMap
}

func isEnd(r io.ReadSeeker) (bool, error) {
	var end uint64
	err := binary.Read(r, binary.BigEndian, &end)
	if err != nil {
		return false, err
	}
	if end == 0 {
		return true, nil
	}
	_, err = r.Seek(-8, 1)
	return false, err
}
