package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
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
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}
	if args.Origin == "" {
		return nil, nil, errors.Newv("missing arg: origin", map[string]interface{}{"args": args})
	}
	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}

	origin, err := zfs.GetDataset(args.Origin)
	if err != nil {
		return nil, nil, err
	}

	ds, err := origin.Clone(args.Name, args.Properties)
	if err != nil {
		return nil, nil, err
	}

	return &DatasetResult{newDataset(ds)}, nil, nil
}
