package nv

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeStructGood(t *testing.T) {
	for i, test := range good {
		t.Log(test.name)

		structs, ok := goodTargetStructs[test.name]
		if !ok {
			fmt.Println("skipping:", test.name)
			continue
		}
		v := reflect.New(reflect.TypeOf(structs.ptr).Elem())
		err := DecodeStruct(test.payload, v.Interface())
		if err != nil {
			t.Log(test.name, "failed:", err)
			t.Fail()
		}

		maps, ok := goodTargetMaps[test.name]
		if !ok {
			fmt.Println("skipping:", test.name)
			continue
		}
		m := reflect.New(reflect.TypeOf(maps.ptr).Elem())
		err = DecodeStruct(test.payload, m.Interface())
		if err != nil {
			t.Log(test.name, "failed:", err)
			t.Fail()
		}

		e, _ := json.Marshal(structs.out)
		o, _ := json.Marshal(v.Elem().Interface())
		ms, _ := json.Marshal(m.Elem().Interface())
		t.Logf("Results:\nexp: %s\nstr: %s\nmap: %s\n", string(e), string(o), string(ms))
		assert.True(t, assert.ObjectsAreEqualValues(structs.out, v.Elem().Interface()), test.name)

		if i == len(goodTargetStructs) {
			return
		}
	}
}
