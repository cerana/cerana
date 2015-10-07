package nv

import (
	"math"
	"time"
)

type decodeTest struct {
	ptr interface{}
	out interface{}
}

type Bools struct {
	True  bool `nv:"true"`
	False bool `nv:"false"`
}
type BoolsMap map[string]bool

type BoolArray struct {
	True  []bool `nv:"true,true,true,true,true"`
	False []bool `nv:"false,false,false,false,false"`
}
type BoolArrayMap map[string][]bool

type Bytes struct {
	A byte `nv:"-128"`
	B byte `nv:"0"`
	C byte `nv:"1"`
	D byte `nv:"127"`
}
type BytesMap map[string]byte

type ByteArray struct {
	A []byte `nv:"-128,-128,-128,-128,-128"`
	B []byte `nv:"0,0,0,0,0"`
	C []byte `nv:"1,1,1,1,1"`
	D []byte `nv:"127,127,127,127,127"`
}
type ByteArrayMap map[string][]byte

type Int8s struct {
	A int8 `nv:"-128"`
	B int8 `nv:"-127"`
	C int8 `nv:"-64"`
	D int8 `nv:"-1"`
	E int8 `nv:"0"`
	F int8 `nv:"1"`
	G int8 `nv:"63"`
	H int8 `nv:"126"`
	I int8 `nv:"127"`
}
type Int8sMap map[string]int8

type Int16s struct {
	A int16 `nv:"-32768"`
	B int16 `nv:"-32767"`
	C int16 `nv:"-16384"`
	D int16 `nv:"-1"`
	E int16 `nv:"0"`
	F int16 `nv:"1"`
	G int16 `nv:"16383"`
	H int16 `nv:"32766"`
	I int16 `nv:"32767"`
}
type Int16sMap map[string]int16

type Int32s struct {
	A int32 `nv:"-2147483648"`
	B int32 `nv:"-2147483647"`
	C int32 `nv:"-1073741824"`
	D int32 `nv:"-1"`
	E int32 `nv:"0"`
	F int32 `nv:"1"`
	G int32 `nv:"1073741823"`
	H int32 `nv:"2147483646"`
	I int32 `nv:"2147483647"`
}
type Int32sMap map[string]int32

type Int64s struct {
	A int64 `nv:"-9223372036854775808"`
	B int64 `nv:"-9223372036854775807"`
	C int64 `nv:"-4611686018427387904"`
	D int64 `nv:"-1"`
	E int64 `nv:"0"`
	F int64 `nv:"1"`
	G int64 `nv:"4611686018427387903"`
	H int64 `nv:"9223372036854775806"`
	I int64 `nv:"9223372036854775807"`
}
type Int64sMap map[string]int64

type Strings struct {
	A string `nv:"0"`
	B string `nv:"1"`
	C string `nv:"HiMom"`
	D string `nv:"\xff\"; DROP TABLE USERS;"`
}
type StringsMap map[string]string

type HRTimes struct {
	A time.Duration `nv:"-9223372036854775808"`
	B time.Duration `nv:"-9223372036854775807"`
	C time.Duration `nv:"-4611686018427387904"`
	D time.Duration `nv:"-1"`
	E time.Duration `nv:"0"`
	F time.Duration `nv:"1"`
	G time.Duration `nv:"4611686018427387903"`
	H time.Duration `nv:"9223372036854775806"`
	I time.Duration `nv:"9223372036854775807"`
}
type HRTimesMap map[string]time.Duration

type NVList struct {
	Nested Bools `nv:"nvlist"`
}
type NVListMap map[string]BoolsMap

type NVListArray struct {
	List []Bools `nv:"list,list"`
}
type NVListArrayMap map[string][]BoolsMap

var goodTargetStructs = map[string]decodeTest{
	"empty":      {ptr: new(struct{}), out: struct{}{}},
	"bools":      {ptr: new(Bools), out: Bools{True: true, False: false}},
	"bool array": {ptr: new(BoolArray), out: BoolArray{True: []bool{true, true, true, true, true}, False: []bool{false, false, false, false, false}}},
	"bytes":      {ptr: new(Bytes), out: Bytes{A: 128, B: 0, C: 1, D: 127}},
	"byte array": {ptr: new(ByteArray), out: ByteArray{A: []byte{128, 128, 128, 128, 128}, B: []byte{0, 0, 0, 0, 0}, C: []byte{1, 1, 1, 1, 1}, D: []byte{127, 127, 127, 127, 127}}},
	"int8s":      {ptr: new(Int8s), out: Int8s{A: math.MinInt8, B: math.MinInt8 + 1, C: math.MinInt8 / 2, D: -1, E: 0, F: 1, G: math.MaxInt8 / 2, H: math.MaxInt8 - 1, I: math.MaxInt8}},
	"int16s":     {ptr: new(Int16s), out: Int16s{A: math.MinInt16, B: math.MinInt16 + 1, C: math.MinInt16 / 2, D: -1, E: 0, F: 1, G: math.MaxInt16 / 2, H: math.MaxInt16 - 1, I: math.MaxInt16}},
	"int32s":     {ptr: new(Int32s), out: Int32s{A: math.MinInt32, B: math.MinInt32 + 1, C: math.MinInt32 / 2, D: -1, E: 0, F: 1, G: math.MaxInt32 / 2, H: math.MaxInt32 - 1, I: math.MaxInt32}},
	"int64s":     {ptr: new(Int64s), out: Int64s{A: math.MinInt64, B: math.MinInt64 + 1, C: math.MinInt64 / 2, D: -1, E: 0, F: 1, G: math.MaxInt64 / 2, H: math.MaxInt64 - 1, I: math.MaxInt64}},
	"strings":    {ptr: new(Strings), out: Strings{A: "0", B: "1", C: "HiMom", D: "\xff\"; DROP TABLE USERS;"}},
	"hrtimes":    {ptr: new(HRTimes), out: HRTimes{A: time.Duration(math.MinInt64), B: time.Duration(math.MinInt64 + 1), C: time.Duration(math.MinInt64 / 2), D: time.Duration(-1), E: time.Duration(0), F: time.Duration(1), G: time.Duration(math.MaxInt64 / 2), H: time.Duration(math.MaxInt64 - 1), I: time.Duration(math.MaxInt64)}},

	"nvlist":       {ptr: new(NVList), out: NVList{Nested: Bools{True: true, False: false}}},
	"nvlist array": {ptr: new(NVListArray), out: NVListArray{List: []Bools{{True: true, False: false}, {True: true, False: false}}}},
}

var goodTargetMaps = map[string]decodeTest{
	"empty":      {ptr: new(struct{}), out: struct{}{}},
	"bools":      {ptr: new(BoolsMap), out: BoolsMap{"true": true, "false": false}},
	"bool array": {ptr: new(BoolArrayMap), out: BoolArrayMap{"true,true,true,true,true": []bool{true, true, true, true, true}, "false,false,false,false,false": []bool{false, false, false, false, false}}},
	"bytes":      {ptr: new(BytesMap), out: BytesMap{"-128": 128, "0": 0, "1": 1, "127": 127}},
	"byte array": {ptr: new(ByteArrayMap), out: ByteArrayMap{"-128,-128,-128,-128,-128": []byte{128, 128, 128, 128, 128}, "0,0,0,0,0": []byte{0, 0, 0, 0, 0}, "1,1,1,1,1": []byte{1, 1, 1, 1, 1}, "127,127,127,127,127": []byte{127, 127, 127, 127, 127}}},
	"int8s":      {ptr: new(Int8sMap), out: Int8sMap{"-128": math.MinInt8, "-127": math.MinInt8 + 1, "-64": math.MinInt8 / 2, "-1": -1, "0": 0, "1": 1, "63": math.MaxInt8 / 2, "126": math.MaxInt8 - 1, "127": math.MaxInt8}},
	"int16s":     {ptr: new(Int16sMap), out: Int16sMap{"-32768": math.MinInt16, "-32767": math.MinInt16 + 1, "-16384": math.MinInt16 / 2, "-1": -1, "0": 0, "1": 1, "16383": math.MaxInt16 / 2, "32766": math.MaxInt16 - 1, "32767": math.MaxInt16}},
	"int32s":     {ptr: new(Int32sMap), out: Int32sMap{"-2147483648": math.MinInt32, "-2147483647": math.MinInt32 + 1, "-1073741824": math.MinInt32 / 2, "-1": -1, "0": 0, "1": 1, "1073741823": math.MaxInt32 / 2, "2147483646": math.MaxInt32 - 1, "2147483647": math.MaxInt32}},
	"int64s":     {ptr: new(Int64sMap), out: Int64sMap{"-9223372036854775808": math.MinInt64, "-9223372036854775807": math.MinInt64 + 1, "-4611686018427387904": math.MinInt64 / 2, "-1": -1, "0": 0, "1": 1, "4611686018427387903": math.MaxInt64 / 2, "9223372036854775806": math.MaxInt64 - 1, "9223372036854775807": math.MaxInt64}},
	"strings":    {ptr: new(StringsMap), out: StringsMap{"0": "0", "1": "1", "HiMom": "HiMom", "\xff\"; DROP TABLE USERS;": "\xff\"; DROP TABLE USERS;"}},
	"hrtimes":    {ptr: new(HRTimesMap), out: HRTimesMap{"-9223372036854775808": time.Duration(math.MinInt64), "-9223372036854775807": time.Duration(math.MinInt64 + 1), "-4611686018427387904": time.Duration(math.MinInt64 / 2), "-1": time.Duration(-1), "0": time.Duration(0), "1": time.Duration(1), "4611686018427387903": time.Duration(math.MaxInt64 / 2), "9223372036854775806": time.Duration(math.MaxInt64 - 1), "9223372036854775807": time.Duration(math.MaxInt64)}},

	"nvlist":       {ptr: new(NVListMap), out: NVListMap{"nvlist": BoolsMap{"true": true, "false": false}}},
	"nvlist array": {ptr: new(NVListArrayMap), out: NVListArrayMap{"list,list": []BoolsMap{{"true": true, "false": false}, {"true": true, "false": false}}}},
}
