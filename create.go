package main

import "github.com/mistifyio/gozfs/nv"

func create(name string, createType dmuType, props map[string]interface{}) error {
	m := map[string]interface{}{
		"cmd":     "zfs_create",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"type":  createType,
			"props": props,
		},
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded, nil)
}
