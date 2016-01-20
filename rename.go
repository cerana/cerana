package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

func rename(name, newName string, recursive bool) (string, error) {
	m := map[string]interface{}{
		"cmd":     "zfs_rename",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"newname":   newName,
			"recursive": recursive,
		},
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return "", err
	}

	var failedName string
	out := make([]byte, 1024)
	err = ioctl(zfs, name, encoded, out)
	if err != nil && recursive {
		_ = nv.NewXDRDecoder(bytes.NewReader(out)).Decode(&failedName)
	}
	return failedName, err
}
