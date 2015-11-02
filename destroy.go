package main

import "github.com/mistifyio/gozfs/nv"

func destroy(name string, deferFlag bool) error {
	m := map[string]interface{}{
		"cmd":     "zfs_destroy",
		"version": uint64(0),
		"defer":   deferFlag,
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded, nil)
}
