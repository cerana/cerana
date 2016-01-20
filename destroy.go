package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

func destroy(name string, deferFlag bool) error {
	m := map[string]interface{}{
		"cmd":     "zfs_destroy",
		"version": uint64(0),
		"defer":   deferFlag,
	}

	encoded := &bytes.Buffer{}
	err := nv.NewXDREncoder(encoded).Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded.Bytes(), nil)
}
