package nv

import (
	"bytes"
	"fmt"
	"io"
	"sort"
)

//go:generate stringer -type=flag
type flag uint32

const (
	_ flag = iota
	_UNIQUE_NAME
	_UNIQUE_NAME_TYPE
)

//go:generate stringer -type=dataType
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
	err := Decode(src, &l)
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
	_UNKNOWN dataType = iota
	_BOOLEAN
	_BYTE
	_INT16
	_UINT16
	_INT32
	_UINT32
	_INT64
	_UINT64
	_STRING
	_BYTE_ARRAY
	_INT16_ARRAY
	_UINT16_ARRAY
	_INT32_ARRAY
	_UINT32_ARRAY
	_INT64_ARRAY
	_UINT64_ARRAY // 0x10
	_STRING_ARRAY
	_HRTIME
	_NVLIST
	_NVLIST_ARRAY
	_BOOLEAN_VALUE
	_INT8
	_UINT8
	_BOOLEAN_ARRAY
	_INT8_ARRAY
	_UINT8_ARRAY
	_DOUBLE
)

type encoding struct {
	Encoding  uint8
	Endianess uint8
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
	case _BOOLEAN:
		valSize = 0
	case _BYTE, _INT8, _UINT8, _INT16, _UINT16, _BOOLEAN_VALUE, _INT32, _UINT32:
		valSize = 4
	case _INT64, _UINT64, _HRTIME, _DOUBLE:
		valSize = 8
	case _BYTE_ARRAY:
		valSize = int(p.NElements * 1)
	case _BOOLEAN_ARRAY, _INT8_ARRAY, _UINT8_ARRAY, _INT16_ARRAY, _UINT16_ARRAY, _INT32_ARRAY, _UINT32_ARRAY:
		valSize = 4 + int(p.NElements*4)
	case _INT64_ARRAY, _UINT64_ARRAY:
		valSize = 4 + int(p.NElements*8)
	case _STRING:
		valSize = 4 + len(p.data.(string))
	case _NVLIST, _NVLIST_ARRAY:
		valSize = len(p.data.([]byte))
	case _STRING_ARRAY:
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
	nvpair_tSize := 4 + 2 + 2 + 4 + 4 + len(p.Name) + 1

	valSize := 0
	switch p.Type {
	case _BOOLEAN:
		valSize = 0
	case _BYTE, _INT8, _UINT8:
		valSize = 1
	case _INT16, _UINT16:
		valSize = 2
	case _BOOLEAN_VALUE, _INT32, _UINT32:
		valSize = 4
	case _INT64, _UINT64, _HRTIME, _DOUBLE:
		valSize = 8
	case _BYTE_ARRAY, _INT8_ARRAY, _UINT8_ARRAY:
		valSize = int(p.NElements * 1)
	case _INT16_ARRAY, _UINT16_ARRAY:
		valSize = int(p.NElements * 2)
	case _INT32_ARRAY, _UINT32_ARRAY:
		valSize = int(p.NElements * 4)
	case _INT64_ARRAY, _UINT64_ARRAY:
		valSize = int(p.NElements * 8)
	case _STRING:
		valSize = len(p.data.(string)) + 1
	case _NVLIST:
		// /* nvlist header */
		// typedef struct nvlist {
		// 	int32_t		nvl_version;
		// 	uint32_t	nvl_nvflag;	/* persistent flags */
		// 	uint64_t	nvl_priv;	/* ptr to private data if not packed */
		// 	uint32_t	nvl_flag;
		// 	int32_t		nvl_pad;	/* currently not used, for alignment */
		// } nvlist_t;
		valSize = 4 + 4 + 8 + 4 + 4
	case _NVLIST_ARRAY:
		// value_sz = (uint64_t)nelem * sizeof (uint64_t) +
		//	      (uint64_t)nelem * NV_ALIGN(sizeof (nvlist_t));
		valSize = int(p.NElements) * (8 + align8(4+4+8+4+4))
	case _BOOLEAN_ARRAY:
		valSize = 4 + int(p.NElements*4)
	case _STRING_ARRAY:
		valSize = int(p.NElements * 8)
		slice := p.data.([]string)
		for i := range slice {
			valSize += len(slice[i]) + 1
		}
	}
	return align8(nvpair_tSize) + align8(valSize)
}
