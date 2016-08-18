package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
)

// Get returns information about a dataset.
func (z *ZFS) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	ds, err := zfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return &DatasetResult{newDataset(ds)}, nil, nil
}
