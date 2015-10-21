package main

import "gozfs/nv"

func exists(name string) error {
	m := map[string]interface{}{
		"cmd":     "zfs_exists",
		"version": uint64(0),
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		panic(err)
	}

	return ioctl(zfs, name, encoded, nil)
}
