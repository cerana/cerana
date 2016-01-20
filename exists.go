package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

func exists(name string) error {
	m := map[string]interface{}{
		"cmd":     "zfs_exists",
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err := nv.Encode(encoded, m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded.Bytes(), nil)
}
