package zfs

import (
	"errors"
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
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
		return nil, nil, errors.New("missing arg: name")
	}

	exists, err := gozfs.Exists(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return &ExistsResult{exists}, nil, nil
}
