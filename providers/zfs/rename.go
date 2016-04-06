package zfs

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/zfs"
)

// RenameArgs are arguments for the Rename handler.
type RenameArgs struct {
	Name      string `json:"name"`
	Origin    string `json:"origin"`
	Recursive bool   `json:"recursive"`
}

// Rename will create a rename from a snapshot.
func (z *ZFS) Rename(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RenameArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if args.Origin == "" {
		return nil, nil, errors.New("missing arg: origin")
	}

	origin, err := zfs.GetDataset(args.Origin)
	if err != nil {
		return nil, nil, err
	}

	_, err = origin.Rename(args.Name, args.Recursive)
	return nil, nil, err
}
