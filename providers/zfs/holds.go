package zfs

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/zfs"
)

// HoldsResult is the result data for the Holds handler.
type HoldsResult struct {
	Holds []string `json:"holds"`
}

// Holds retrieves a list of user holds on the specified snapshot.
func (z *ZFS) Holds(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	ds, err := zfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	holds, err := ds.Holds()
	if err != nil {
		return nil, nil, err
	}

	return &HoldsResult{holds}, nil, nil
}
