package main

import "github.com/mistifyio/gozfs/nv"

func clone(name, origin string, props map[string]interface{}) error {
	m := map[string]interface{}{
		"cmd":     "zfs_clone",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"origin": origin,
			"props":  props,
		},
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded, nil)
}
