package nv

import "reflect"

type decoder interface {
	Decode(interface{}) error
	header() (header, error)
	meta() (string, dataType, error)
	value(reflect.Type) (reflect.Value, fieldSetFunc, error)
	skip() error
	isEnd() (bool, error)
}
