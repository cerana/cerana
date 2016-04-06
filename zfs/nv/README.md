# nv

[![nv](https://godoc.org/github.com/cerana/cerana/zfs/nv?status.png)](https://godoc.org/github.com/cerana/cerana/zfs/nv)



## Usage

#### func  NewNativeEncoder

```go
func NewNativeEncoder(w io.Writer) *nativeEncoder
```

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


#### func  NewNativeDecoder

```go
func NewNativeDecoder(r io.ReadSeeker) *NativeDecoder
```

#### func (*NativeDecoder) Decode

```go
func (d *NativeDecoder) Decode(target interface{}) (err error)
```
Decode Note: care should be taken when decoding into a `map[string]interface{}`
as bytes/uint8s (and their array forms) can not be distinguished and will be
treated as uint8/[]uint8.

#### type XDRDecoder

```go
type XDRDecoder struct {
	*xdr.Decoder
}
```


#### func  NewXDRDecoder

```go
func NewXDRDecoder(r io.ReadSeeker) *XDRDecoder
```

#### func (*XDRDecoder) Decode

```go
func (d *XDRDecoder) Decode(target interface{}) error
```
Decode Note: care should be taken when decoding into a `map[string]interface{}`
as bytes/uint8s (and their array forms) can not be distinguished and will be
treated as uint8/[]uint8.

#### type XDREncoder

```go
type XDREncoder struct {
}
```


#### func  NewXDREncoder

```go
func NewXDREncoder(w io.Writer) XDREncoder
```

#### func (XDREncoder) Encode

```go
func (e XDREncoder) Encode(i interface{}) error
```

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
