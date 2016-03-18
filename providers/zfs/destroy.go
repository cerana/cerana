package zfs

import (
	"errors"
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// DestroyArgs are arguments for the Destroy handler.
type DestroyArgs struct {
	Name            string `json:"name"`
	Defer           bool   `json:"defer"`
	Recursive       bool   `json:"recursive"`
	RecursiveClones bool   `json:"recursive-clones"`
	ForceUnmount    bool   `json:"force-unmount"`
}

// Destroy will destroy a dataset.
func (z *ZFS) Destroy(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DestroyArgs
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

	opts := &gozfs.DestroyOptions{
		Recursive:       args.Recursive,
		RecursiveClones: args.RecursiveClones,
		ForceUnmount:    args.ForceUnmount,
		Defer:           args.Defer,
	}

	return nil, nil, ds.Destroy(opts)
}
