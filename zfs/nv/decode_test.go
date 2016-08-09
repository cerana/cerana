package nv

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cerana/cerana/pkg/errors"
)

func checkArray(t *testing.T, field string, value interface{}, l int, fn func([]string)) {
	empty := strings.HasPrefix(field, "empty ")

	expecteds := strings.Split(field, ";")
	if !empty {
		if len(expecteds) < 2 {
			t.Fatal("field does not seem to be a list of expected values:", field)
		}
		if len(expecteds) != l {
			t.Fatal("length mismatch between expected and decoded arrays, expected:",
				len(expecteds), "decoded:", l)
		}
		fn(expecteds)
	} else {
		if len(expecteds) != 1 {
			t.Fatal("field seems to be a list of expected values:", field)
		}
		if l != 0 {
			t.Fatal("length mismatch between expected and decoded arrays, expected: 0 decoded:", l)
		}
	}
}

func checkBoolean(t *testing.T, field string, value interface{}) {
	exp := false
	_, err := fmt.Sscanf(field, "%t", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := bool(value.(Boolean))
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkBooleanValue(t *testing.T, field string, value interface{}) {
	exp := false
	_, err := fmt.Sscanf(field, "%t", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(bool)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkBooleanArray(t *testing.T, field string, value interface{}) {
	arr := value.([]bool)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkBooleanValue(t, expecteds[i], arr[i])
		}
	})
}

func checkByte(t *testing.T, field string, value interface{}) {
	d := int(0)
	_, err := fmt.Sscanf(field, "%d", &d)
	if err != nil {
		t.Fatal(err)
	}
	exp := byte(d)
	got := value.(byte)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkByteArray(t *testing.T, field string, value interface{}) {
	arr := value.([]byte)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkByte(t, expecteds[i], arr[i])
		}
	})
}

func checkDouble(t *testing.T, field string, value interface{}) {
	exp := float64(0)
	// is a hack, but good enough
	_, err := fmt.Sscanf(field, "%f", &exp)
	if err != nil {
		t.Fatal(err)
	}

	got := value.(float64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}

func checkHrtime(t *testing.T, field string, value interface{}) {
	d := int64(0)
	_, err := fmt.Sscanf(field, "%d", &d)
	if err != nil {
		t.Fatal(err)
	}
	exp := time.Duration(d)
	got := value.(time.Duration)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}

func checkInt8(t *testing.T, field string, value interface{}) {
	exp := int8(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(int8)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkInt8Array(t *testing.T, field string, value interface{}) {
	arr := value.([]int8)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkInt8(t, expecteds[i], arr[i])
		}
	})
}

func checkInt16(t *testing.T, field string, value interface{}) {
	exp := int16(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(int16)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkInt16Array(t *testing.T, field string, value interface{}) {
	arr := value.([]int16)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkInt16(t, expecteds[i], arr[i])
		}
	})
}

func checkInt32(t *testing.T, field string, value interface{}) {
	exp := int32(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(int32)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkInt32Array(t *testing.T, field string, value interface{}) {
	arr := value.([]int32)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkInt32(t, expecteds[i], arr[i])
		}
	})
}

func checkInt64(t *testing.T, field string, value interface{}) {
	exp := int64(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(int64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkInt64Array(t *testing.T, field string, value interface{}) {
	arr := value.([]int64)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkInt64(t, expecteds[i], arr[i])
		}
	})
}

func checkUint8(t *testing.T, field string, value interface{}) {
	exp := uint8(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(uint8)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkUint8Array(t *testing.T, field string, value interface{}) {
	arr := value.([]uint8)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkUint8(t, expecteds[i], arr[i])
		}
	})
}

func checkUint16(t *testing.T, field string, value interface{}) {
	exp := uint16(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(uint16)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkUint16Array(t *testing.T, field string, value interface{}) {
	arr := value.([]uint16)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkUint16(t, expecteds[i], arr[i])
		}
	})
}

func checkUint32(t *testing.T, field string, value interface{}) {
	exp := uint32(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(uint32)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkUint32Array(t *testing.T, field string, value interface{}) {
	arr := value.([]uint32)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkUint32(t, expecteds[i], arr[i])
		}
	})
}

func checkUint64(t *testing.T, field string, value interface{}) {
	exp := uint64(0)
	_, err := fmt.Sscanf(field, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := value.(uint64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkUint64Array(t *testing.T, field string, value interface{}) {
	arr := value.([]uint64)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkUint64(t, expecteds[i], arr[i])
		}
	})
}

func checkNvlist(t *testing.T, field string, value interface{}) {
	num, err := strconv.Atoi(field)
	if err != nil {
		t.Fatal(err)
	}
	list := value.(map[string]interface{})
	if num != len(list) {
		t.Fatal("length mismatch between expected and decoded lists",
			"expected:", num, "got:", len(list))
	}
	for f, v := range list {
		checkers[reflect.TypeOf(v).String()](t, f, v)
	}
}
func checkNvlistArray(t *testing.T, field string, value interface{}) {
	arr := value.([]map[string]interface{})
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkNvlist(t, expecteds[i], arr[i])
		}
	})
}

func checkString(t *testing.T, field string, value interface{}) {
	exp := field
	got := value.(string)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func checkStringArray(t *testing.T, field string, value interface{}) {
	arr := value.([]string)
	checkArray(t, field, value, len(arr), func(expecteds []string) {
		for i := range arr {
			checkString(t, expecteds[i], arr[i])
		}
	})
}

var checkers map[string]func(*testing.T, string, interface{})

func init() {
	checkers = map[string]func(*testing.T, string, interface{}){
		"nv.Boolean":                checkBoolean,
		"bool":                      checkBooleanValue,
		"byte":                      checkByte,
		"int16":                     checkInt16,
		"uint16":                    checkUint16,
		"int32":                     checkInt32,
		"uint32":                    checkUint32,
		"int64":                     checkInt64,
		"uint64":                    checkUint64,
		"string":                    checkString,
		"[]byte":                    checkByteArray,
		"[]int16":                   checkInt16Array,
		"[]uint16":                  checkUint16Array,
		"[]int32":                   checkInt32Array,
		"[]uint32":                  checkUint32Array,
		"[]int64":                   checkInt64Array,
		"[]uint64":                  checkUint64Array,
		"[]string":                  checkStringArray,
		"time.Duration":             checkHrtime,
		"map[string]interface {}":   checkNvlist,
		"[]map[string]interface {}": checkNvlistArray,
		"int8":    checkInt8,
		"uint8":   checkUint8,
		"[]bool":  checkBooleanArray,
		"[]int8":  checkInt8Array,
		"[]uint8": checkUint8Array,
		"float64": checkDouble,
	}
}

func struct2map(s interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	val := reflect.ValueOf(s).Elem()
	indexMap, _ := fieldIndexMap(val)
	for name, index := range indexMap {
		m[name] = val.Field(index).Interface()
	}
	return m
}

func assertFields(t *testing.T, name string, m map[string]interface{}) {
	for field, value := range m {
		valueType := reflect.TypeOf(value).String()
		fn, ok := checkers[valueType]
		if !ok {
			t.Fatal(name, "unknown type:", valueType)
		}
		fn(t, field, value)
	}
}

type tDecoder struct {
	r io.ReadSeeker
	decoder
}

func decode(t *testing.T, name string, ptr interface{}, dec tDecoder) {
	m := map[string]interface{}{}

	if _, err := dec.r.Seek(0, 0); err != nil {
		t.Fatal(name, "decode seek 0 failed:", err)
	}
	if err := dec.Decode(&m); err != nil {
		t.Fatal(name, "decode as map failed:", err)
	}

	assertFields(t, name, m)

	if _, err := dec.r.Seek(0, 0); err != nil {
		t.Fatal(name, "decode seek 0 failed:", err)
	}
	if err := dec.Decode(ptr); err != nil {
		t.Fatal(name, "decode as struct failed:", err)
	}

	m = struct2map(ptr)
	if len(m) != reflect.ValueOf(ptr).Elem().NumField() {
		t.Fatalf("incorrect number of fields, got: %d %+v want: %d %+v\n",
			len(m), m, reflect.ValueOf(ptr).Elem().NumField(), ptr)
	}

	assertFields(t, name, m)
}

//go:generate make -s -C _test_data ../known_good_data_test.go
func TestDecodeGood(t *testing.T) {
	for _, test := range good {
		if testing.Verbose() {
			fmt.Println(" -- ", test.name)
		} else {
			t.Log(" -- ", test.name)
		}

		r := bytes.NewReader(test.native)
		dec := tDecoder{
			r:       r,
			decoder: NewNativeDecoder(r),
		}
		decode(t, test.name, test.ptr(), dec)

		if test.name == "empty arrays" {
			continue
		}

		r = bytes.NewReader(test.xdr)
		dec = tDecoder{
			r:       r,
			decoder: NewXDRDecoder(r),
		}
		decode(t, test.name, test.ptr(), dec)
	}
}

func TestDecodeBad(t *testing.T) {
	for _, test := range decodeBad {
		if testing.Verbose() {
			fmt.Println(" -- ", test.err)
		} else {
			t.Log(" -- ", test.err)
		}

		m := map[string]interface{}{}
		err := NewXDRDecoder(bytes.NewReader(test.payload)).Decode(&m)
		if err == nil {
			t.Fatalf("expected an error, wanted:|%s| payload:|%v|\n",
				test.err, test.payload)
		}
		if test.err != errors.Cause(err).Error() {
			t.Fatalf("error mismatch, want:|%s| got:|%s| payload:|%v|\n",
				test.err, errors.Cause(err).Error(), test.payload)
		}
	}
}

func TestDecodeBadArgs(t *testing.T) {
	badArgs := []struct {
		arg interface{}
		err string
	}{
		{
			arg: struct{}{},
			err: "cannot decode into non-pointer",
		},
	}
	for _, test := range badArgs {
		err := NewXDRDecoder(bytes.NewReader([]byte(encDecNameTyp))).Decode(test.arg)
		if err == nil {
			t.Fatalf("expected an error, wanted:|%s|\n", test.err)
		}
		if test.err != err.Error() {
			t.Fatalf("error mismatch, want:|%s| got:|%s|\n",
				test.err, err.Error())
		}
	}
}
