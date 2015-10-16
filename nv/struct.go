package nv

import (
	"encoding/binary"
	"io"
	"reflect"
)

//go:generate stringer -type=flag
type flag uint32

const (
	_ flag = iota
	UNIQUE_NAME
	UNIQUE_NAME_TYPE
)

//go:generate stringer -type=dataType
type dataType uint32

const (
	UNKNOWN dataType = iota
	BOOLEAN
	BYTE
	INT16
	UINT16
	INT32
	UINT32
	INT64
	UINT64
	STRING
	BYTE_ARRAY
	INT16_ARRAY
	UINT16_ARRAY
	INT32_ARRAY
	UINT32_ARRAY
	INT64_ARRAY
	UINT64_ARRAY // 0x10
	STRING_ARRAY
	HRTIME
	NVLIST
	NVLIST_ARRAY
	BOOLEAN_VALUE
	INT8
	UINT8
	BOOLEAN_ARRAY
	INT8_ARRAY
	UINT8_ARRAY
	DOUBLE
)

var goToNV = map[string]dataType{
	"bool":          BOOLEAN_VALUE,
	"boolean_array": BOOLEAN_ARRAY,
	"byte":          BYTE,
	"byte_array":    BYTE_ARRAY,
	"float64":       DOUBLE,
	"int16":         INT16,
	"int16_array":   INT16_ARRAY,
	"int32":         INT32,
	"int32_array":   INT32_ARRAY,
	"int64":         INT64,
	"int64_array":   INT64_ARRAY,
	"int8":          INT8,
	"int8_array":    INT8_ARRAY,
	"nv.List":       NVLIST,
	"nvlist_array":  NVLIST_ARRAY,
	"string":        STRING,
	"string_array":  STRING_ARRAY,
	"time.Time":     HRTIME,
	"uint16":        UINT16,
	"uint16_array":  UINT16_ARRAY,
	"uint32":        UINT32,
	"uint32_array":  UINT32_ARRAY,
	"uint64":        UINT64,
	"uint64_array":  UINT64_ARRAY,
	"uint8":         UINT8,
	"uint8_array":   UINT8_ARRAY,
}

type List struct {
	header
	Pairs []Pair
}

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

type Pair struct {
	pair
	Value interface{}
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
	case BYTE, INT8, UINT8, INT16, UINT16, BOOLEAN_VALUE, INT32, UINT32:
		valSize = 4
	case INT64, UINT64, HRTIME, DOUBLE:
		valSize = 8
	case BYTE_ARRAY:
		valSize = int(p.NElements * 1)
	case BOOLEAN_ARRAY, INT8_ARRAY, UINT8_ARRAY, INT16_ARRAY, UINT16_ARRAY, INT32_ARRAY, UINT32_ARRAY:
		valSize = 4 + int(p.NElements*4)
	case INT64_ARRAY, UINT64_ARRAY:
		valSize = 4 + int(p.NElements*8)
	case STRING:
		valSize = 4 + len(p.data.(string))
	case NVLIST, NVLIST_ARRAY:
		valSize = len(p.data.([]byte))
	case STRING_ARRAY:
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
	case BYTE, INT8, UINT8:
		valSize = 1
	case INT16, UINT16:
		valSize = 2
	case BOOLEAN_VALUE, INT32, UINT32:
		valSize = 4
	case INT64, UINT64, HRTIME, DOUBLE:
		valSize = 8
	case BYTE_ARRAY, INT8_ARRAY, UINT8_ARRAY:
		valSize = int(p.NElements * 1)
	case INT16_ARRAY, UINT16_ARRAY:
		valSize = int(p.NElements * 2)
	case INT32_ARRAY, UINT32_ARRAY:
		valSize = int(p.NElements * 4)
	case INT64_ARRAY, UINT64_ARRAY:
		valSize = int(p.NElements * 8)
	case STRING:
		valSize = len(p.data.(string)) + 1
	case NVLIST:
		// /* nvlist header */
		// typedef struct nvlist {
		// 	int32_t		nvl_version;
		// 	uint32_t	nvl_nvflag;	/* persistent flags */
		// 	uint64_t	nvl_priv;	/* ptr to private data if not packed */
		// 	uint32_t	nvl_flag;
		// 	int32_t		nvl_pad;	/* currently not used, for alignment */
		// } nvlist_t;
		valSize = 4 + 4 + 8 + 4 + 4
	case NVLIST_ARRAY:
		// value_sz = (uint64_t)nelem * sizeof (uint64_t) +
		//	      (uint64_t)nelem * NV_ALIGN(sizeof (nvlist_t));
		valSize = int(p.NElements) * (8 + align8(4+4+8+4+4))
	case BOOLEAN_ARRAY:
		valSize = 4 + int(p.NElements*4)
	case STRING_ARRAY:
		valSize = int(p.NElements * 8)
		slice := p.data.([]string)
		for i := range slice {
			valSize += len(slice[i]) + 1
		}
	}
	return align8(nvpair_tSize) + align8(valSize)
}

func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func writeHeader(w io.Writer) error {

	return nil
}

type mVal struct {
	Name  string
	Type  dataType
	Value interface{}
}
type mList map[string]mVal

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
