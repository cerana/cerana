package zfs

import (
	"errors"
	"net/url"
	"strings"
	"syscall"

	"github.com/mistifyio/go-zfs"
	"github.com/mistifyio/mistify/acomm"
)

// TODO: Update this method once `gozfs` supports Unmount

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
		return nil, nil, errors.New("missing arg: name")
	}

	ds, err := zfs.GetDataset(args.Name)
	if err != nil {
		// Fix errors to be more like what gozfs will probably return
		if strings.Contains(err.Error(), "dataset does not exist") {
			err = syscall.ENOENT
		}
		return nil, nil, err
	}

	ds, err = ds.Unmount(args.Force)
	if err != nil {
		// Fix errors to be more like what gozfs will probably return
		if strings.Contains(err.Error(), "not currently mounted") {
			err = syscall.EINVAL
		}
		return nil, nil, err
	}

	return nil, nil, nil
}
