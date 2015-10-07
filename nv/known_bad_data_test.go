package nv

const enc_dec = "\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x01\x00\x00\x00\x01"

//                                  name length          ┌╴name...fill    beginning of type
const enc_dec_name_typ = enc_dec + "\x00\x00\x00\x01" + "0\x00\x00\x00" + "\x00\x00\x00"

var decode_bad = []struct {
	err     string
	payload []byte
}{
	{"EOF", []byte{}},
	{"unexpected EOF", []byte("\x00")},
	{"EOF", []byte("\x00\x00\x00\x00")},
	{"invalid encoding: 2", []byte("\x02\x00\x00\x00")},
	{"invalid endianess: 2", []byte("\x01\x02\x00\x00")},
	{"unexpected reserved1 value: 1", []byte("\x01\x01\x01\x00")},
	{"unexpected reserved2 value: 1", []byte("\x01\x01\x00\x01")},
	{"unexpected EOF", []byte("\x01\x01\x00\x00\x00\x00\x00\x00")},
	{"unexpected version: 1", []byte("\x01\x01\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unexpected Flag: flag(0)", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unexpected Flag: flag(3)", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x03")},
	{"unexpected EOF", []byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00")},
	{"unknown type: UNKNOWN", []byte(enc_dec_name_typ + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unknown type: dataType(28)", []byte(enc_dec_name_typ + string(byte(DOUBLE+1)) + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"EOF", []byte(enc_dec_name_typ + string(byte(BYTE)) + "\x00\x00\x00\x01\x00\x00\x00\x00")},
	{"unexpected EOF", []byte(enc_dec_name_typ + string(byte(BYTE)) + "\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"xdr:DecodeUint: EOF while decoding 4 bytes - read: '[]'", []byte(enc_dec_name_typ + string(byte(BYTE)) + "\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01")},
	{"xdr:DecodeString: data exceeds max slice limit - read: '4294967295'", []byte(enc_dec_name_typ + string(byte(STRING)) + "\x00\x00\x00\x00\xff\xff\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unknown type: UNKNOWN", []byte(enc_dec_name_typ + string(byte(NVLIST)) + "\x00\x00\x00\x01" + string([]byte(enc_dec_name_typ + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")[4:]) + "\x00\x00\x00\x00\x00\x00\x00\x00")},
	{"unknown type: UNKNOWN", []byte(enc_dec_name_typ + string(byte(NVLIST_ARRAY)) + "\x00\x00\x00\x01" + string([]byte(enc_dec_name_typ + "\x00" + "\x00\x00\x00\x01\x00\x00\x00\x00")[4:]) + "\x00\x00\x00\x00\x00\x00\x00\x00")},
}

var encode_bad = []struct {
	err     string
	payload interface{}
}{
	{"can not encode a nil pointer", nil},
	{"invalid type 'int', must be a struct", 0},
	{"invalid Type 'UNKNOWN'", mList{"foo1": mVal{}}},
	{"invalid Type 'dataType(28)'", mList{"foo2": mVal{Type: DOUBLE + 1}}},
	{"xdr:Marshal: can't marshal nil interface", mList{"foo2": mVal{Type: DOUBLE}}},
	//{"empty NVLIST_ARRAY", List{Pairs: []Pair{Pair{pair: pair{Type: NVLIST_ARRAY}, Value: []List{}}}}},
}
