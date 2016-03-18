package zfs

import (
	"net/url"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
)

// ListArgs are the args for the List handler.
type ListArgs struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}

// ListResult is the result data for the List handler.
type ListResult struct {
	Datasets []*Dataset `json:"datasets"`
}

// List returns a list of filesystems, volumes, snapshots, and bookmarks
func (z *ZFS) List(req *acomm.Request) (interface{}, *url.URL, error) {
	var args ListArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if len(args.Types) == 0 {
		args.Types = []string{"all"}
	}

	list, err := gozfs.Datasets(args.Name, args.Types)
	if err != nil {
		return nil, nil, err
	}

	result := &ListResult{
		Datasets: make([]*Dataset, len(list)),
	}
	for i, ds := range list {
		result.Datasets[i] = newDataset(ds)
	}

	return result, nil, nil
}
