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

// Decode
// Note: care should be taken when decoding into a `map[string]interface{}` as
// bytes/uint8s (and their array forms) can not be distinguished and will be
// treated as uint8/[]uint8.
func Decode(data []byte, target interface{}) (err error) {
	b := bytes.NewReader(data)

	// Validate data encoding
	enc := encoding{}
	if err := binary.Read(b, binary.BigEndian, &enc); err != nil {
		return err
	}
	if enc.Encoding > 1 {
		return fmt.Errorf("invalid encoding: %v", enc.Encoding)
	}
	if enc.Endianess > 1 {
		return fmt.Errorf("invalid endianess: %v", enc.Endianess)
	}
	if enc.Reserved1 != 0 {
		return fmt.Errorf("unexpected reserved1 value: %v", enc.Reserved1)
	}
	if enc.Reserved2 != 0 {
		return fmt.Errorf("unexpected reserved2 value: %v", enc.Reserved2)
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

func decodeList(r io.ReadSeeker, target reflect.Value) error {
	// Validate data header
	var h header
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
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
		if _, err := xdr.Unmarshal(r, &dataPair); err != nil {
			return err
		}

		// If not dealing with a map, look up the corresponding struct field
		var targetField reflect.Value
		if !isMap {
			targetFieldIndex, ok := targetFieldIndexMap[dataPair.Name]
			// If there's no corresponding struct field, skip the data and move
			// on to the next data pair
			if !ok {
				r.Seek(int64(dataPair.EncodedSize-uint32(dataPair.headerSize())), 1)
				continue
			} else {
				targetField = target.Field(targetFieldIndex)
			}
		}

		// val used to set if target is a map
		var val reflect.Value
		// fieldSetFunc used to set if target is a struct field
		var fieldSetFunc func()

		var targetType reflect.Type
		if isList {
			targetType = target.Type()
		} else if isMap {
			targetType = target.Type().Elem()
		} else {
			targetType = targetField.Type()
		}

		var err error
		var v interface{}
		dec := newXDRDecoder(r)
		switch dataPair.Type {
		case _BOOLEAN:
			v := Boolean(true)
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _BOOLEAN_VALUE:
			v, err = dec.decodeBool()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetBool(v.(bool))
			}
		case _BYTE:
			v, err = dec.decodeByte()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint8)))
			}
		case _INT8:
			v, err = dec.decodeInt8()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int8)))
			}
		case _INT16:
			v, err = dec.decodeInt16()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int16)))
			}
		case _INT32:
			v, err = dec.decodeInt32()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int32)))
			}
		case _INT64:
			v, err = dec.decodeInt64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(v.(int64))
			}
		case _UINT8:
			v, err = dec.decodeUint8()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint8)))
			}
		case _UINT16:
			v, err = dec.decodeUint16()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint16)))
			}
		case _UINT32:
			v, err = dec.decodeUint32()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint32)))
			}
		case _UINT64:
			v, err = dec.decodeUint64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint64)))
			}
		case _HRTIME:
			v, err = dec.decodeHRTime()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(time.Duration)))
			}
		case _DOUBLE:
			v, err = dec.decodeFloat64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetFloat(v.(float64))
			}
		case _BOOLEAN_ARRAY:
			v, err = dec.decodeBoolArray()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _BYTE_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.decodeByteArray()
				val = reflect.ValueOf(v)
				fieldSetFunc = func() {
					targetField.SetBytes(v.([]byte))
				}
			}
		case _INT8_ARRAY:
			v, err = dec.decodeInt8Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _INT16_ARRAY:
			v, err = dec.decodeInt16Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _INT32_ARRAY:
			v, err = dec.decodeInt32Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _INT64_ARRAY:
			v, err = dec.decodeInt64Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _UINT8_ARRAY:
			v, err = dec.decodeUint8Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _UINT16_ARRAY:
			v, err = dec.decodeUint16Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _UINT32_ARRAY:
			v, err = dec.decodeUint32Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _UINT64_ARRAY:
			v, err = dec.decodeUint64Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _STRING:
			v, err = dec.decodeString()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetString(v.(string))
			}
		case _STRING_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.decodeStringArray()
				val = reflect.ValueOf(v)
				fieldSetFunc = func() {
					targetField.Set(val)
				}
			}
		case _NVLIST:
			val = reflect.Indirect(reflect.New(targetType))
			err = decodeList(r, val)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case _NVLIST_ARRAY:
			if targetType.Kind() == reflect.Interface {
				targetType = reflect.TypeOf([]map[string]interface{}{})
			}

			val = reflect.MakeSlice(targetType, 0, int(dataPair.NElements))
			for i := uint32(0); i < dataPair.NElements; i++ {
				elem := reflect.Indirect(reflect.New(targetType.Elem()))
				err = decodeList(r, elem)
				if err != nil {
					break
				}
				val = reflect.Append(val, elem)
			}
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		default:
			return fmt.Errorf("unknown type: %v", dataPair.Type)
		}

		if err != nil {
			return err
		}

		// Set the value appropriately
		if isMap {
			name := reflect.ValueOf(dataPair.Name)
			if isList {
				val = reflect.ValueOf(map[string]interface{}{
					"type":  dataPair.Type.String(),
					"value": val.Interface(),
				})
			}
			target.SetMapIndex(name, val)
		} else {
			fieldSetFunc()
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
