package nv

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
)

func diff(want, got []byte) (string, string) {
	w := strings.Split(fmt.Sprintf("%03v", want), " ")
	w[0] = strings.TrimLeft(w[0], "[")
	w[len(w)-1] = strings.TrimRight(w[len(w)-1], "]")

	g := strings.Split(fmt.Sprintf("%03v", got), " ")
	g[0] = strings.TrimLeft(g[0], "[")
	g[len(g)-1] = strings.TrimRight(g[len(g)-1], "]")

	min := len(w)
	if len(g) < min {
		min = len(g)
	}
	red := "\x1b[31;1m"
	green := "\x1b[32;1m"
	reset := "\x1b[0m"
	diff := false
	for i := 0; i < min; i++ {
		if g[i] != w[i] {
			if diff == false {
				diff = true
				g[i] = red + g[i]
				w[i] = green + w[i]
			}
		} else if diff {
			diff = false
			g[i] = reset + g[i]
			w[i] = reset + w[i]
		}
	}
	if len(g) > min {
		diff = true
		g[min] = red + g[min]
	}
	if len(w) > min {
		diff = true
		w[min] = green + w[min]
	}

	if diff {
		g[len(g)-1] += reset
		w[len(w)-1] += reset
	}
	w[0] = "[" + w[0]
	w[len(w)-1] += "]"
	g[0] = "[" + g[0]
	g[len(g)-1] += "]"
	return strings.Join(w, " "), strings.Join(g, " ")
}

func assertEqual(t *testing.T, name, typ string, payload []byte, enc tEncoder, i interface{}) {
	err := enc.Encode(i)
	if err != nil {
		t.Fatalf("%s: failed to encode as %s: error:%s\n", name, typ, err)
	}
	encoded := enc.w.Bytes()
	if !reflect.DeepEqual(payload, encoded) {
		want, got := diff(payload, encoded)
		t.Fatalf("%s: %s: encoded does not match payload\nwant:|%s|\n got:|%s|\n",
			name, typ, want, got)
	}
}

type tEncoder struct {
	w *bytes.Buffer
	encoder
}

func encode(t *testing.T, name string, data []byte, ptr interface{}, dec tDecoder, enc tEncoder) {
	m := map[string]interface{}{}

	dec.r.Seek(0, 0)
	if err := dec.Decode(&m); err != nil {
		t.Fatal(name, "decode as map failed:", err)
	}

	if !strings.Contains(name, "byte") {
		assertEqual(t, name, "map", data, enc, m)
	}
	enc.w.Reset()

	dec.r.Seek(0, 0)
	if err := dec.Decode(ptr); err != nil {
		t.Fatal(name, "decode as struct failed:", err)
	}

	assertEqual(t, name, "struct", data, enc, ptr)
	enc.w.Reset()
}

func TestEncodeGood(t *testing.T) {
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
		w := &bytes.Buffer{}
		enc := tEncoder{
			w:       w,
			encoder: NewNativeEncoder(w),
		}
		encode(t, test.name, test.native, test.ptr(), dec, enc)

		if test.name == "empty arrays" {
			continue
		}

		r = bytes.NewReader(test.xdr)
		dec = tDecoder{
			r:       r,
			decoder: NewXDRDecoder(r),
		}
		w = &bytes.Buffer{}
		enc = tEncoder{
			w:       w,
			encoder: NewXDREncoder(w),
		}
		encode(t, test.name, test.xdr, test.ptr(), dec, enc)
	}
}

func TestEncodeBad(t *testing.T) {
	for _, test := range encode_bad {
		if testing.Verbose() {
			fmt.Println(" -- ", test.err)
		} else {
			t.Log(" -- ", test.err)
		}

		enc := NewXDREncoder(ioutil.Discard)
		err := enc.Encode(test.payload)
		if err == nil {
			t.Fatalf("expected an error, wanted:|%s|, payload:|%v|\n",
				test.err, test.payload)
		}
		if test.err != err.Error() {
			t.Fatalf("error mismatch, want:|%s|, got:|%s|, payload:|%v|\n",
				test.err, err.Error(), test.payload)
		}
	}
}
