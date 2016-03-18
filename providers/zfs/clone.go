package zfs

import (
	"errors"
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// CloneArgs are arguments for the Clone handler.
type CloneArgs struct {
	Name       string                 `json:"name"`
	Origin     string                 `json:"origin"`
	Properties map[string]interface{} `json:"properties"`
}

// Clone will create a clone from a snapshot.
func (z *ZFS) Clone(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CloneArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if args.Origin == "" {
		return nil, nil, errors.New("missing arg: origin")
	}
	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}

	origin, err := gozfs.GetDataset(args.Origin)
	if err != nil {
		return nil, nil, err
	}

	ds, err := origin.Clone(args.Name, args.Properties)
	if err != nil {
		return nil, nil, err
	}

	return &DatasetResult{newDataset(ds)}, nil, nil
}
