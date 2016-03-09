package zfs

import (
	"errors"
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// RollbackArgs are the arguments for the Rollback handler.
type RollbackArgs struct {
	Name          string `json:"name"`
	DestroyRecent bool   `json:"destroy-recent"`
}

// Rollback rolls a filesystem or volume back to a given snapshot.
func (z *ZFS) Rollback(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RollbackArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	ds, err := gozfs.GetDataset(args.Name)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, ds.Rollback(args.DestroyRecent)
}
