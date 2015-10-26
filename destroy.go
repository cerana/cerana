package main

import "github.com/mistifyio/gozfs/nv"

func destroy(name string) error {
	m := map[string]interface{}{
		"cmd":     "zfs_destroy",
		"version": uint64(0),
		"defer":   false,
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		panic(err)
	}

	return ioctl(zfs, name, encoded, nil)
}
