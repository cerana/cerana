package nv

import (
	"io"

	xdr "github.com/davecgh/go-xdr/xdr2"
)

type xdrDecoder struct {
	*xdr.Decoder
}

func newXDRDecoder(r io.Reader) xdrDecoder {
	return xdrDecoder{Decoder: xdr.NewDecoder(r)}
}
