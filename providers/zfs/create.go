package zfs

import (
	"errors"
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// CreateArgs are arguments for the Create handler.
type CreateArgs struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // gozfs.Dataset[Filesystem,Snapshot,Volume]
	Volsize    uint64                 `json:"volsize"`
	Properties map[string]interface{} `json:"properties"`
}

// CreateResult is the result from the Create Handler.
type CreateResult struct {
	Dataset *Dataset
}

// Create will create a new filesystem or volume dataset.
func (z *ZFS) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}

	var ds *gozfs.Dataset
	var err error
	switch args.Type {
	case gozfs.DatasetFilesystem:
		ds, err = gozfs.CreateFilesystem(args.Name, args.Properties)
	case gozfs.DatasetVolume:
		if args.Volsize <= 0 {
			err = errors.New("missing or invalid arg: volsize")
			break
		}
		ds, err = gozfs.CreateVolume(args.Name, args.Volsize, args.Properties)
	default:
		err = errors.New("missing or invalid arg: type")
	}

	if err != nil {
		return nil, nil, err
	}

	return &CreateResult{newDataset(ds)}, nil, nil
}
