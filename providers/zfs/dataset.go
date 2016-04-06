package zfs

import (
	"path/filepath"

	"github.com/cerana/cerana/zfs"
)

// Dataset contains information and properties for a ZFS dataset. This struct
// is the same as zfs.Dataset, except all methods that interact with ZFS have
// been removed. The ZFS provider should be the only place that interacts with
// zfs directly.
//
// Use this struct for datasets anywhere outside the ZFS provider.
type Dataset struct {
	Name           string
	Properties     *zfs.DatasetProperties
	DMUObjsetStats *zfs.DMUObjsetStats
}

// Mountpoint returns the resolved mountpoint of the dataset.
func (d *Dataset) Mountpoint() string {
	// dataset mountpoint property joined to the dataset name with the
	// mountpointsource property trimmed from the name.
	defaultPart, _ := filepath.Rel(d.Properties.MountpointSource, d.Name)
	return filepath.Join(d.Properties.Mountpoint, defaultPart)
}

func newDataset(orig *zfs.Dataset) *Dataset {
	if orig == nil {
		return nil
	}

	return &Dataset{
		Name:           orig.Name,
		Properties:     orig.Properties,
		DMUObjsetStats: orig.DMUObjsetStats,
	}
}
