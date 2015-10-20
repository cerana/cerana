package nv

import (
	"reflect"
	"strings"
)

// getTags returns the content of the "nv" tag after splitting on ','
func getTags(i int, v reflect.Value) []string {
	return strings.Split(v.Type().Field(i).Tag.Get("nv"), ",")
}

// forEachField calls `fn` as it iterates through each field in struct `v`
func forEachField(v reflect.Value, fn func(int, reflect.Value)) {
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fn(i, field)
	}
}
