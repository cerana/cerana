package nv

import (
	"reflect"
	"strings"
)

// getTags returns the content of the "nv" tag after splitting on ','
func getTags(i int, v reflect.Value) []string {
	return strings.Split(v.Type().Field(i).Tag.Get("nv"), ",")
}

// forEachField calls `fn` as it iterates through each field in struct `v` until
// there are no more fields or `fn` returns false
func forEachField(v reflect.Value, fn func(int, reflect.Value) bool) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if keepGoing := fn(i, field); !keepGoing {
			return
		}
	}
}

// deref will follow all pointers until it reaches a value, which it returns
func deref(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
