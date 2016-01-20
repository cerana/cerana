package nv

type encoder interface {
	Encode(interface{}) error
	header(header) error
	item(string, dataType, interface{}) error
	footer() error
}
