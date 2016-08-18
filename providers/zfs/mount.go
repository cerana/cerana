package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
)

// MountArgs are arguments for the Mount handler.
type MountArgs struct {
	Name    string `json:"name"`
	Overlay bool   `json:"overlay"`
}

// Mount mounts a zfs filesystem.
func (z *ZFS) Mount(req *acomm.Request) (interface{}, *url.URL, error) {
	var args MountArgs
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

	return nil, nil, ds.Mount(args.Overlay, nil)
}
