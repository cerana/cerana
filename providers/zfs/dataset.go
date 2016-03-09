package zfs

import (
	"path/filepath"

	"github.com/mistifyio/gozfs"
)

// Dataset contains information and properties for a ZFS dataset. This struct
// is the same as gozfs.Dataset, except all methods that interact with ZFS have
// been removed. The ZFS provider should be the only place that interacts with
// zfs directly.
//
// Use this struct for datasets anywhere outside the ZFS provider.
type Dataset struct {
	Name           string
	Properties     *gozfs.DatasetProperties
	DMUObjsetStats *gozfs.DMUObjsetStats
}

// Mountpoint returns the mountpoint of the dataset. It is based off of the
// dataset mountpoint property joined to the dataset name with the
// mountpointsource property trimmed from the name.
func (d *Dataset) Mountpoint() string {
	defaultPart, _ := filepath.Rel(d.Properties.MountpointSource, d.Name)
	return filepath.Join(d.Properties.Mountpoint, defaultPart)
}

func newDataset(orig *gozfs.Dataset) *Dataset {
	if orig == nil {
		return nil
	}

	return &Dataset{
		Name:           orig.Name,
		Properties:     orig.Properties,
		DMUObjsetStats: orig.DMUObjsetStats,
	}
}
