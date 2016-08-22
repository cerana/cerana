package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
)

// UnmountArgs are arguments for the Unmount handler.
type UnmountArgs struct {
	Name  string `json:"name"`
	Force bool   `json:"force"`
}

// Unmount mounts a zfs filesystem.
func (z *ZFS) Unmount(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UnmountArgs
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

	return nil, nil, ds.Unmount(args.Force)
}
