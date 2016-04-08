# nv

[![nv](https://godoc.org/github.com/cerana/cerana/zfs/nv?status.png)](https://godoc.org/github.com/cerana/cerana/zfs/nv)



## Usage

#### func  PrettyPrint

```go
func PrettyPrint(dst *bytes.Buffer, src []byte, indent string) error
```
PrettyPrint writes into `dst` a decoded form of `src` meant for human
consumption. The nv/xdr types are printed in sorted order along with the name
and value of the nvp.

#### type Boolean

```go
type Boolean bool
```

Boolean is used for encoding and decoding the nvlist BOOLEAN datatype, which is
different than BOOLEAN_VALUE. An nvlist BOOLEAN data type consists of a string
key and a zero length value. Instead of dealing with inconsistencies when
working with go structs vs map[string]interface, special go property
annotations, and/or inferring when to handle data as a BOOLEAN based on
nil/interface{} values, this explicitly named go type is to be used.

#### type NativeDecoder

```go
type NativeDecoder struct {
}
```

NativeDecoder is a Decoder for native encoding.

#### func  NewNativeDecoder

```go
func NewNativeDecoder(r io.ReadSeeker) *NativeDecoder
```
NewNativeDecoder creates a new NativeDecoder.

#### func (*NativeDecoder) Decode

```go
func (d *NativeDecoder) Decode(target interface{}) (err error)
```
Decode decodes data into a supplied target. Note: care should be taken when
decoding into a `map[string]interface{}` as bytes/uint8s (and their array forms)
can not be distinguished and will be treated as uint8/[]uint8.

#### type NativeEncoder

```go
type NativeEncoder struct {
}
```

NativeEncoder is an Encoder for native encoding.

#### func  NewNativeEncoder

```go
func NewNativeEncoder(w io.Writer) *NativeEncoder
```
NewNativeEncoder creates a new nativeEncoder.

#### func (NativeEncoder) Encode

```go
func (e NativeEncoder) Encode(i interface{}) error
```
Encode encodes the supplied data with native encoding.

#### type XDRDecoder

```go
type XDRDecoder struct {
	*xdr.Decoder
}
```

XDRDecoder is a Decoder for XDR encoding.

#### func  NewXDRDecoder

```go
func NewXDRDecoder(r io.ReadSeeker) *XDRDecoder
```
NewXDRDecoder creates a new XDRDecoder.

#### func (*XDRDecoder) Decode

```go
func (d *XDRDecoder) Decode(target interface{}) error
```
Decode decodes data into a supplied target. Note: care should be taken when
decoding into a `map[string]interface{}` as bytes/uint8s (and their array forms)
can not be distinguished and will be treated as uint8/[]uint8.

#### type XDREncoder

```go
type XDREncoder struct {
}
```

XDREncoder is an Encoder for XDR encoding.

#### func  NewXDREncoder

```go
func NewXDREncoder(w io.Writer) XDREncoder
```
NewXDREncoder creates a new XDREncoder.

#### func (XDREncoder) Encode

```go
func (e XDREncoder) Encode(i interface{}) error
```
Encode encodes data using XDR encoding.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
