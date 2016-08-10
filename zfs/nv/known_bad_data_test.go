package nv

const encDec = "\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01"

//                                  name length          ┌╴name...fill    beginning of type
const encDecNameTyp = encDec + "\x00\x00\x00\x01" + "0\x00\x00\x00" + "\x00\x00\x00"

var decodeBad = []struct {
	err     string
	payload []byte
}{
	{"EOF", []byte{}},
	{"unexpected EOF", []byte("\x00")},
	{"EOF", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01")},
	{"invalid encoding", []byte("\x02\x00\x00\x00")},
	{"invalid endianness", []byte("\x01\x02\x00\x00")},
	{"unexpected reserved1 value", []byte("\x01\x01\x01\x00")},
	{"unexpected reserved2 value", []byte("\x01\x01\x00\x01")},
	{"unexpected EOF", []byte("\x01\x01\x00\x00\x00\x00\x00\x00")},
	{"unexpected version", []byte("\x01\x01\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unexpected flag", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unexpected flag", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03")},
	{"unexpected EOF", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00")},
	{"unknown type", []byte(encDecNameTyp + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unknown type", []byte(encDecNameTyp + string(byte(_double+1)) + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"EOF", []byte(encDecNameTyp + string(byte(_byte)) + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unexpected EOF", []byte(encDecNameTyp + string(byte(_byte)) + "\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"xdr:DecodeUint: EOF while decoding 4 bytes - read: '[]'", []byte(encDecNameTyp + string(byte(_byte)) + "\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01")},
	{"xdr:DecodeString: data exceeds max slice limit - read: '4294967295'", []byte(encDecNameTyp + string(byte(_string)) + "\x00\x00\x00\x00\xff\xff\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unknown type", []byte(encDecNameTyp + string(byte(_nvlist)) + "\x00\x00\x00\x01" + string([]byte(encDecNameTyp + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")[4:]) + "\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unknown type", []byte(encDecNameTyp + string(byte(_nvlistArray)) + "\x00\x00\x00\x01" + string([]byte(encDecNameTyp + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")[4:]) + "\x00\x00\x00\x00\x00\x00\x00\x00")},
}

var encodeBad = []struct {
	err     string
	payload interface{}
}{
	{"can not encode a nil pointer", nil},
	{"invalid type, must be a struct", 0},
}
