package nv

import (
	"bytes"
	"fmt"
	"io"
	"sort"
)

//go:generate sh -c "stringer -type=flag && sed -i 's#_flag_name#flagName#g;s#_flag_index#flagIndex#g' flag_string.go"
type flag uint32

const (
	_ flag = iota
	uniqueName
	uniqueNameType
)

//go:generate sh -c "stringer -type=dataType && sed -i 's#_dataType_name#dataTypeName#g;s#_dataType_index#dataTypeIndex#g' datatype_string.go"
type dataType uint32

// Boolean is used for encoding and decoding the nvlist BOOLEAN datatype, which
// is different than BOOLEAN_VALUE.  An nvlist BOOLEAN data type consists of a
// string key and a zero length value. Instead of dealing with inconsistencies
// when working with go structs vs map[string]interface, special go property
// annotations, and/or inferring when to handle data as a BOOLEAN based on
// nil/interface{} values, this explicitly named go type is to be used.
type Boolean bool

// PrettyPrint writes into `dst` a decoded form of `src` meant for human
// consumption. The nv/xdr types are printed in sorted order along with the name
// and value of the nvp.
func PrettyPrint(dst *bytes.Buffer, src []byte, indent string) error {
	m, err := pretty(src)
	if err != nil {
		return err
	}
	prettyPrint(dst, m, indent, "")
	return nil
}

type list map[string]interface{}

// pretty decodes the NVList in src as a `nv.list` which stores the nv type
// information.
func pretty(src []byte) (map[string]interface{}, error) {

	l := list{}
	err := NewXDRDecoder(bytes.NewReader(src)).Decode(&l)
	return l, err
}

func prettyPrint(dst io.Writer, m map[string]interface{}, indenter, indent string) {
	names := make([]string, 0, len(m))
	for k := range m {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		v := m[name]
		value, ok := v.(map[string]interface{})
		if !ok {
			fmt.Fprintf(dst, "%sName: %s, Value:%+v\n", indent, name, v)
			continue
		}

		if value["type"] == "_NVLIST" {
			fmt.Fprintf(dst, "%sName: %s, Type: %s\n", indent, name, value["type"])
			prettyPrint(dst, value["value"].(list), indenter, indent+indenter)
		} else {
			fmt.Fprintf(dst, "%sName: %s, Type: %s, Value:%+v\n",
				indent, name, value["type"], value["value"])
		}
	}
}

const (
	_unknown dataType = iota
	_boolean
	_byte
	_int16
	_uint16
	_int32
	_uint32
	_int64
	_uint64
	_string
	_byteArray
	_int16Array
	_uint16Array
	_int32Array
	_uint32Array
	_int64Array
	_uint64Array // 0x10
	_stringArray
	_hrtime
	_nvlist
	_nvlistArray
	_booleanValue
	_int8
	_uint8
	_booleanArray
	_int8Array
	_uint8Array
	_double
)

//go:generate sh -c "stringer -type=codec && sed -i 's#_codec_name#codecName#g;s#_codec_index#codecIndex#g' codec_string.go"
type codec uint8

const (
	nativeCodec codec = iota
	xdrCodec
	maxCodec = xdrCodec
)

//go:generate sh -c "stringer -type=endianness && sed -i 's#_endianness_name#endiannessName#g;s#_endianness_index#endiannessIndex#g' endianness_string.go"
type endianness uint8

const (
	bigEndian endianness = iota
	littleEndian
	maxEndian = littleEndian
)

type encoding struct {
	Encoding  codec
	Endianess endianness
	Reserved1 uint8
	Reserved2 uint8
}

type header struct {
	Version uint32
	Flag    flag
}

type pair struct {
	EncodedSize uint32
	DecodedSize uint32
	Name        string
	Type        dataType
	NElements   uint32
	data        interface{}
}

func align4(n int) int {
	return (n + 3) & ^3
}

func align8(n int) int {
	return (n + 7) & ^7
}

func (p pair) headerSize() int {
	return 4 + 4 + 4 + align4(len(p.Name)) + 4 + 4
}

func (p pair) encodedSize() int {
	valSize := 0
	switch p.Type {
	case _boolean:
		valSize = 0
	case _byte, _int8, _uint8, _int16, _uint16, _booleanValue, _int32, _uint32:
		valSize = 4
	case _int64, _uint64, _hrtime, _double:
		valSize = 8
	case _byteArray:
		valSize = int(p.NElements * 1)
	case _booleanArray, _int8Array, _uint8Array, _int16Array, _uint16Array, _int32Array, _uint32Array:
		valSize = 4 + int(p.NElements*4)
	case _int64Array, _uint64Array:
		valSize = 4 + int(p.NElements*8)
	case _string:
		valSize = 4 + len(p.data.(string))
	case _nvlist, _nvlistArray:
		valSize = len(p.data.([]byte))
	case _stringArray:
		slice := p.data.([]string)
		for i := range slice {
			valSize += align4(4 + len(slice[i]))
		}
	}
	return p.headerSize() + align4(valSize)
}

func (p pair) decodedSize() int {
	// typedef struct nvpair {
	// 	int32_t nvp_size;	/* size of this nvpair */
	// 	int16_t	nvp_name_sz;	/* length of name string */
	// 	int16_t	nvp_reserve;	/* not used */
	// 	int32_t	nvp_value_elem;	/* number of elements for array types */
	// 	data_type_t nvp_type;	/* type of value */
	// 	/* name string */
	// 	/* aligned ptr array for string arrays */
	// 	/* aligned array of data for value */
	// } nvpair_t;
	nvpairTSize := 4 + 2 + 2 + 4 + 4 + len(p.Name) + 1

	valSize := 0
	switch p.Type {
	case _boolean:
		valSize = 0
	case _byte, _int8, _uint8:
		valSize = 1
	case _int16, _uint16:
		valSize = 2
	case _booleanValue, _int32, _uint32:
		valSize = 4
	case _int64, _uint64, _hrtime, _double:
		valSize = 8
	case _byteArray, _int8Array, _uint8Array:
		valSize = int(p.NElements * 1)
	case _int16Array, _uint16Array:
		valSize = int(p.NElements * 2)
	case _int32Array, _uint32Array:
		valSize = int(p.NElements * 4)
	case _int64Array, _uint64Array:
		valSize = int(p.NElements * 8)
	case _string:
		valSize = len(p.data.(string)) + 1
	case _nvlist:
		// /* nvlist header */
		// typedef struct nvlist {
		// 	int32_t		nvl_version;
		// 	uint32_t	nvl_nvflag;	/* persistent flags */
		// 	uint64_t	nvl_priv;	/* ptr to private data if not packed */
		// 	uint32_t	nvl_flag;
		// 	int32_t		nvl_pad;	/* currently not used, for alignment */
		// } nvlist_t;
		valSize = 4 + 4 + 8 + 4 + 4
	case _nvlistArray:
		// value_sz = (uint64_t)nelem * sizeof (uint64_t) +
		//	      (uint64_t)nelem * NV_ALIGN(sizeof (nvlist_t));
		valSize = int(p.NElements) * (8 + align8(4+4+8+4+4))
	case _booleanArray:
		valSize = 4 + int(p.NElements*4)
	case _stringArray:
		valSize = int(p.NElements * 8)
		slice := p.data.([]string)
		for i := range slice {
			valSize += len(slice[i]) + 1
		}
	}
	return align8(nvpairTSize) + align8(valSize)
}
