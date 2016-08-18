package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
)

// ExistsResult is the result data for the Exists handler.
type ExistsResult struct {
	Exists bool `json:"exists"`
}

// Exists determines whether a dataset exists or not.
func (z *ZFS) Exists(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	exists, err := zfs.Exists(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return &ExistsResult{exists}, nil, nil
}
