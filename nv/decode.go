package nv

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
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

// fieldSetFunc is used to set the value if the target is a field within a
// struct
type fieldSetFunc func(reflect.Value, reflect.Value)

func decodeList(dec decoder, target reflect.Value) error {
	// Validate data header
	h, err := dec.header()
	if err != nil {
		return err
	}
	if h.Version != 0 {
		return fmt.Errorf("unexpected version: %v", h.Version)
	}
	if h.Flag < _UNIQUE_NAME || h.Flag > _UNIQUE_NAME_TYPE {
		return fmt.Errorf("unexpected Flag: %v", h.Flag)
	}

	// If the target is a pointer, initialize and dereference.
	if target.Kind() == reflect.Ptr {
		target.Set(reflect.New(target.Type().Elem()))
		target = reflect.Indirect(target)
	}

	// maps and structs have slightly different handling
	// interface types will be handled as maps
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
		// Done when there's no more data or an error has occurred
		if end, err := dec.isEnd(); end || err != nil {
			return err
		}

		// Get next encoded data pair
		// Note: This just gets the data pair metadata. The actual value
		// data is left on the reader, to be pulled off by the decoder.
		pName, pType, err := dec.meta()
		if err != nil {
			return err
		}

		// If not dealing with a map, look up the corresponding struct field
		var targetField reflect.Value
		if !isMap {
			targetFieldIndex, ok := targetFieldIndexMap[pName]
			// If there's no corresponding struct field, skip the data and move
			// on to the next data pair
			if !ok {
				if err := dec.skip(); err != nil {
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
		value, fieldSetter, err := dec.value(targetType)
		if err != nil {
			return err
		}

		// Set the value appropriately
		if isMap {
			name := reflect.ValueOf(pName)
			if isList {
				value = reflect.ValueOf(map[string]interface{}{
					"type":  pType.String(),
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
	forEachField(v, func(i int, field reflect.Value) bool {
		// Skip fields that can't be set (e.g. unexported)
		if !field.CanSet() {
			return true
		}
		name := v.Type().Field(i).Name
		if tags := getTags(i, v); len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}
		vFieldIndexMap[name] = i
		return true
	})
	return vFieldIndexMap
}
