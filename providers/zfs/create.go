package zfs

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs"
)

// CreateArgs are arguments for the Create handler.
type CreateArgs struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // zfs.Dataset[Filesystem,Volume]
	Volsize    uint64                 `json:"volsize"`
	Properties map[string]interface{} `json:"properties"`
}

// Create will create a new filesystem or volume dataset.
func (z *ZFS) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.Newv("missing arg: name", map[string]interface{}{"args": args})
	}

	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}

	var ds *zfs.Dataset
	var err error
	switch args.Type {
	case zfs.DatasetFilesystem:
		ds, err = zfs.CreateFilesystem(args.Name, args.Properties)
	case zfs.DatasetVolume:
		if args.Volsize <= 0 {
			err = errors.Newv("missing or invalid arg: volsize", map[string]interface{}{"args": args})
			break
		}
		ds, err = zfs.CreateVolume(args.Name, args.Volsize, args.Properties)
	default:
		err = errors.Newv("missing or invalid arg: type", map[string]interface{}{"args": args})
	}

	if err != nil {
		return nil, nil, err
	}

	return &DatasetResult{newDataset(ds)}, nil, nil
}
