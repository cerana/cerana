package main

import (
	"syscall"

	"github.com/mistifyio/gozfs/nv"
)

func snapshot(zpool string, snapNames []string, props map[string]string) (map[string]int32, error) {
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
		return nil, err
	}

	var errlist map[string]int32
	out := make([]byte, 1024)
	err = ioctl(zfs, zpool, encoded, out)
	if errno, ok := err.(syscall.Errno); ok && errno == syscall.EEXIST {
		// Try to get errlist info, but ignore any errors in the attempt
		_ = nv.Decode(out, &errlist)
	}
	return errlist, err
}
