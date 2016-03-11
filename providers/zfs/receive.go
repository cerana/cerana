package zfs

import (
	"errors"
	"io"
	"net/url"

	"github.com/Sirupsen/logrus"
	zfs "github.com/mistifyio/go-zfs"
	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify-logrus-ext"
	"github.com/mistifyio/mistify/acomm"
)

// ReceiveResult is the result from the Receive handler
type ReceiveResult struct {
	Dataset *Dataset
}

// TODO: Update this method once `gozfs` supports receive

// Receive creates a new snapshot from a zfs stream. If it a full stream, then
// a new filesystem or volume is created as well.
func (z *ZFS) Receive(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if req.StreamURL == nil {
		return nil, nil, errors.New("missing requet StreamURL")
	}

	r, w := io.Pipe()
	go func() {
		defer logrusx.LogReturnedErr(w.Close, nil, "failed to close writer")
		logrusx.LogReturnedErr(func() error {
			return acomm.Stream(w, req.StreamURL)
		}, logrus.Fields{"streamURL": req.StreamURL}, "failed to stream")
	}()

	if _, err := zfs.ReceiveSnapshot(r, args.Name); err != nil {
		return nil, nil, err
	}

	ds, err := gozfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return &ReceiveResult{newDataset(ds)}, nil, nil
}
