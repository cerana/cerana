package main

import (
	"syscall"

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

	out := make([]byte, 1024)
	err = ioctl(zfs, zpool, encoded, out)
	if errno, ok := err.(syscall.Errno); ok && errno == 17 {
		if err := nv.Decode(out, errlist); err != nil {
			panic(err)
		}
	}
	return err
}
