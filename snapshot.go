package main

import (
	"bytes"
	"syscall"

	"github.com/mistifyio/gozfs/nv"
)

func snapshot(zpool string, snapNames []string, props map[string]string) (map[string]syscall.Errno, error) {
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

	encoded := &bytes.Buffer{}
	err := nv.Encode(encoded, m)
	if err != nil {
		return nil, err
	}

	var errlist map[string]syscall.Errno
	out := make([]byte, 1024)
	err = ioctl(zfs, zpool, encoded.Bytes(), out)
	if errno, ok := err.(syscall.Errno); ok && errno == syscall.EEXIST {
		// Try to get errlist info, but ignore any errors in the attempt
		errs := map[string]int32{}
		if nv.NewXDRDecoder(bytes.NewReader(out)).Decode(&errs) == nil {
			errlist = make(map[string]syscall.Errno, len(errs))
			for snap, errno := range errs {
				errlist[snap] = syscall.Errno(errno)
			}
		}
	}
	return errlist, err
}
