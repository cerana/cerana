package nv

import (
	"reflect"
	"strings"
)

// getTags returns the content of the "nv" tag after splitting on ','
func getTags(i int, v reflect.Value) []string {
	return strings.Split(v.Type().Field(i).Tag.Get("nv"), ",")
}
