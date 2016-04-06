package zfs

import (
	"errors"
	"io"
	"net/url"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	gzfs "github.com/mistifyio/go-zfs"
	"github.com/cerana/cerana/zfs"
	"github.com/mistifyio/mistify-logrus-ext"
)

// TODO: Update this method once `zfs` supports receive

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
		return nil, nil, errors.New("missing request stream-url")
	}

	r, w := io.Pipe()
	go func() {
		defer logrusx.LogReturnedErr(w.Close, nil, "failed to close writer")
		logrusx.LogReturnedErr(func() error {
			return acomm.Stream(w, req.StreamURL)
		}, logrus.Fields{"streamURL": req.StreamURL}, "failed to stream")
	}()

	if _, err := gzfs.ReceiveSnapshot(r, args.Name); err != nil {
		// Fix errors to be more like what zfs will probably return
		if strings.Contains(err.Error(), "dataset does not exist") {
			err = syscall.ENOENT
		} else if strings.Contains(err.Error(), "exists\nmust specify -F to overwrite") {
			err = syscall.EEXIST
		}
		return nil, nil, err
	}

	ds, err := zfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return &DatasetResult{newDataset(ds)}, nil, nil
}
