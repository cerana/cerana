package nv

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"time"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

func decodeOld(buf []byte) (mList, error) {
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

func Decode(data []byte, target interface{}) (err error) {
	// Catch any panics from reflection
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			if rs, ok := r.(string); ok {
				err = errors.New(rs)
			} else {
				err = r.(error)
			}
		}
	}()

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

	return decodeListStruct(b, reflect.Indirect(targetV))
}

func decodeListStruct(r io.ReadSeeker, target reflect.Value) error {
	// Validate data header
	var h header
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return err
	}
	if h.Version != 0 {
		return fmt.Errorf("unexpected version: %v", h.Version)
	}
	if h.Flag < UNIQUE_NAME || h.Flag > UNIQUE_NAME_TYPE {
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

		var err error
		var v interface{}
		dec := newDecoder(r)
		switch dataPair.Type {
		case BOOLEAN_VALUE:
			v, err = dec.DecodeBool()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetBool(v.(bool))
			}
		case BYTE:
			v, err = dec.DecodeByte()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint8)))
			}
		case INT8:
			v, err = dec.DecodeInt8()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int8)))
			}
		case INT16:
			v, err = dec.DecodeInt16()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int16)))
			}
		case INT32:
			v, err = dec.DecodeInt32()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(int32)))
			}
		case INT64:
			v, err = dec.DecodeInt64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(v.(int64))
			}
		case UINT8:
			v, err = dec.DecodeUint8()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint8)))
			}
		case UINT16:
			v, err = dec.DecodeUint16()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint16)))
			}
		case UINT32:
			v, err = dec.DecodeUint32()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint32)))
			}
		case UINT64:
			v, err = dec.DecodeUint64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetUint(uint64(v.(uint64)))
			}
		case HRTIME:
			v, err = dec.DecodeHRTime()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetInt(int64(v.(time.Duration)))
			}
		case DOUBLE:
			v, err = dec.DecodeFloat64()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetFloat(v.(float64))
			}
		case BOOLEAN_ARRAY:
			v, err = dec.DecodeBoolArray()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case BYTE_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.DecodeByteArray()
				val = reflect.ValueOf(v)
				fieldSetFunc = func() {
					targetField.SetBytes(v.([]byte))
				}
			}
		case INT8_ARRAY:
			v, err = dec.DecodeInt8Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case INT16_ARRAY:
			v, err = dec.DecodeInt16Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case INT32_ARRAY:
			v, err = dec.DecodeInt32Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case INT64_ARRAY:
			v, err = dec.DecodeInt64Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case UINT8_ARRAY:
			v, err = dec.DecodeUint8Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case UINT16_ARRAY:
			v, err = dec.DecodeUint16Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case UINT32_ARRAY:
			v, err = dec.DecodeUint32Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case UINT64_ARRAY:
			v, err = dec.DecodeUint64Array()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case STRING:
			v, err = dec.DecodeString()
			val = reflect.ValueOf(v)
			fieldSetFunc = func() {
				targetField.SetString(v.(string))
			}
		case STRING_ARRAY:
			if _, err = r.Seek(-4, 1); err == nil {
				v, err = dec.DecodeStringArray()
				val = reflect.ValueOf(v)
				fieldSetFunc = func() {
					targetField.Set(val)
				}
			}
		case NVLIST:
			if isMap {
				val = reflect.Indirect(reflect.New(target.Type().Elem()))
			} else {
				val = reflect.Indirect(reflect.New(targetField.Type()))
			}
			err = decodeListStruct(r, val)
			fieldSetFunc = func() {
				targetField.Set(val)
			}
		case NVLIST_ARRAY:
			var sliceType reflect.Type
			if isMap {
				sliceType = target.Type().Elem()
			} else {
				sliceType = targetField.Type()
			}

			if sliceType.Kind() == reflect.Interface {
				sliceType = reflect.TypeOf([]map[string]interface{}{})
			}

			val = reflect.MakeSlice(sliceType, 0, int(dataPair.NElements))
			for i := uint32(0); i < dataPair.NElements; i++ {
				elem := reflect.Indirect(reflect.New(sliceType.Elem()))
				err = decodeListStruct(r, elem)
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
			target.SetMapIndex(reflect.ValueOf(dataPair.Name), val)
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
	for i := 0; i < v.NumField(); i++ {
		vField := v.Field(i)
		// Skip fields that can't be set (e.g. unexported)
		if !vField.CanSet() {
			continue
		}
		vTypeField := v.Type().Field(i)
		dataFieldName := vTypeField.Name
		if tags := getTags(i, v); len(tags) > 0 && tags[0] != "" {
			dataFieldName = tags[0]
		}
		vFieldIndexMap[dataFieldName] = i
	}
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
