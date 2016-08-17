package zfs

import (
	"io"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/zfs"
)

// Receive creates a new snapshot from a zfs stream. If it a full stream, then
// a new filesystem or volume is created as well.
func (z *ZFS) Receive(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}
	if req.StreamURL == nil {
		return nil, nil, errors.Newv("missing request stream-url", map[string]interface{}{"args": args})
	}

	r, w := io.Pipe()
	go func() {
		defer logrusx.LogReturnedErr(w.Close, nil, "failed to close writer")
		logrusx.LogReturnedErr(func() error {
			return acomm.Stream(w, req.StreamURL)
		}, map[string]interface{}{"streamURL": req.StreamURL}, "failed to stream")
	}()

	return nil, nil, zfs.Receive(r, args.Name)
}
