package nv

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func check_array(t *testing.T, p mVal, l int, fn func([]string)) {
	expecteds := strings.Split(p.Name, ";")
	if len(expecteds) < 2 {
		t.Fatal("p.Name does not seem to be a list of expected values:", p.Name)
	}
	if len(expecteds) != l {
		t.Fatal("length mismatch between expected and decoded arrays, expected:",
			len(expecteds), "decoded:", l)
	}
	fn(expecteds)
}

func check_boolean(t *testing.T, p mVal) {
	check_boolean_value(t, p)
}
func check_boolean_value(t *testing.T, p mVal) {
	exp := false
	_, err := fmt.Sscanf(p.Name, "%t", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(bool)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_boolean_array(t *testing.T, p mVal) {
	arr := p.Value.([]bool)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_boolean_value(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_byte(t *testing.T, p mVal) {
	d := int(0)
	_, err := fmt.Sscanf(p.Name, "%d", &d)
	if err != nil {
		t.Fatal(err)
	}
	exp := byte(d)
	got := p.Value.(byte)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_byte_array(t *testing.T, p mVal) {
	arr := p.Value.([]byte)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_byte(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_double(t *testing.T, p mVal) {
	exp := float64(0)
	// is a hack, but good enough
	_, err := fmt.Sscanf(p.Name, "%f", &exp)
	if err != nil {
		t.Fatal(err)
	}

	got := p.Value.(float64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}

func check_hrtime(t *testing.T, p mVal) {
	d := int64(0)
	_, err := fmt.Sscanf(p.Name, "%d", &d)
	if err != nil {
		t.Fatal(err)
	}
	exp := time.Duration(d)
	got := p.Value.(time.Duration)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}

func check_int8(t *testing.T, p mVal) {
	exp := int8(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(int8)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_int8_array(t *testing.T, p mVal) {
	arr := p.Value.([]int8)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_int8(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_int16(t *testing.T, p mVal) {
	exp := int16(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(int16)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_int16_array(t *testing.T, p mVal) {
	arr := p.Value.([]int16)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_int16(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_int32(t *testing.T, p mVal) {
	exp := int32(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(int32)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_int32_array(t *testing.T, p mVal) {
	arr := p.Value.([]int32)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_int32(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_int64(t *testing.T, p mVal) {
	exp := int64(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(int64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_int64_array(t *testing.T, p mVal) {
	arr := p.Value.([]int64)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_int64(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_uint8(t *testing.T, p mVal) {
	exp := uint8(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(uint8)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_uint8_array(t *testing.T, p mVal) {
	arr := p.Value.([]uint8)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_uint8(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_uint16(t *testing.T, p mVal) {
	exp := uint16(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(uint16)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_uint16_array(t *testing.T, p mVal) {
	arr := p.Value.([]uint16)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_uint16(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_uint32(t *testing.T, p mVal) {
	exp := uint32(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(uint32)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_uint32_array(t *testing.T, p mVal) {
	arr := p.Value.([]uint32)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_uint32(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_uint64(t *testing.T, p mVal) {
	exp := uint64(0)
	_, err := fmt.Sscanf(p.Name, "%d", &exp)
	if err != nil {
		t.Fatal(err)
	}
	got := p.Value.(uint64)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_uint64_array(t *testing.T, p mVal) {
	arr := p.Value.([]uint64)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_uint64(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_nvlist(t *testing.T, p mVal) {
	list := p.Value.(mList)
	for _, pair := range list {
		checkers[pair.Type](t, pair)
	}
}
func check_nvlist_array(t *testing.T, p mVal) {
	arr := p.Value.([]mList)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_nvlist(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

func check_string(t *testing.T, p mVal) {
	exp := p.Name
	got := p.Value.(string)
	if got != exp {
		t.Fatal("expected:", exp, "got:", got)
	}
}
func check_string_array(t *testing.T, p mVal) {
	arr := p.Value.([]string)
	check_array(t, p, len(arr), func(expecteds []string) {
		for i := range arr {
			check_string(t, mVal{Name: expecteds[i], Value: arr[i]})
		}
	})
}

var checkers map[dataType]func(*testing.T, mVal)

func init() {
	checkers = map[dataType]func(*testing.T, mVal){
		BOOLEAN:       check_boolean,
		BYTE:          check_byte,
		INT16:         check_int16,
		UINT16:        check_uint16,
		INT32:         check_int32,
		UINT32:        check_uint32,
		INT64:         check_int64,
		UINT64:        check_uint64,
		STRING:        check_string,
		BYTE_ARRAY:    check_byte_array,
		INT16_ARRAY:   check_int16_array,
		UINT16_ARRAY:  check_uint16_array,
		INT32_ARRAY:   check_int32_array,
		UINT32_ARRAY:  check_uint32_array,
		INT64_ARRAY:   check_int64_array,
		UINT64_ARRAY:  check_uint64_array,
		STRING_ARRAY:  check_string_array,
		HRTIME:        check_hrtime,
		NVLIST:        check_nvlist,
		NVLIST_ARRAY:  check_nvlist_array,
		BOOLEAN_VALUE: check_boolean_value,
		INT8:          check_int8,
		UINT8:         check_uint8,
		BOOLEAN_ARRAY: check_boolean_array,
		INT8_ARRAY:    check_int8_array,
		UINT8_ARRAY:   check_uint8_array,
		DOUBLE:        check_double,
	}
}

//go:generate make -C _test ../known_good_data_test.go
func TestDecodeGood(t *testing.T) {
	for _, test := range good {
		t.Log(test.name)

		m, err := Decode(test.payload)
		if err != nil {
			t.Fatal(test.name, "failed:", err)
		}

		for _, p := range m {
			fn, ok := checkers[p.Type]
			if !ok {
				t.Fatal(test.name, "unknown type:", p.Type)
			}
			fn(t, p)
		}
	}
}

func TestDecodeBad(t *testing.T) {
	for _, test := range decode_bad {
		t.Log(test.err)

		_, err := Decode(test.payload)
		if err == nil {
			t.Fatalf("expected an error, wanted:|%s| payload:|%v|\n",
				test.err, test.payload)
		}
		if test.err != err.Error() {
			t.Fatalf("error mismatch, want:|%s| got:|%s| payload:|%v|\n",
				test.err, err.Error(), test.payload)
		}
	}
}
