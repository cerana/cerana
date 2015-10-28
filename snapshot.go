package main

import (
	"io"

	"github.com/mistifyio/gozfs/nv"
)

func snapshot(zpool string, snapNames []string, props map[string]string, errlist *map[string]int32) error {
	// snaps needs to be a map with the snap name as the key and an arbitrary value
	snaps := make(map[string]string)
	for _, snapName := range snapNames {
		snaps[snapName] = ""
	}

	m := map[string]interface{}{
		"cmd":     "zfs_snapshot",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"snaps": snaps,
			"props": props,
		},
	}
	encoded, err := nv.Encode(m)
	if err != nil {
		panic(err)
	}

	var out []byte
	err = ioctl(zfs, zpool, encoded, out)
	if err := nv.Decode(out, errlist); err != nil && err != io.EOF {
		panic(err)
	}
	return err
}
