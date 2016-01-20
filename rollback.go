package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

func rollback(name string) (string, error) {
	m := map[string]interface{}{
		"cmd":     "zfs_rollback",
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err := nv.NewXDREncoder(encoded).Encode(m)
	if err != nil {
		return "", err
	}

	var snapName string
	out := make([]byte, 1024)
	err = ioctl(zfs, name, encoded.Bytes(), out)
	if err == nil {
		var results map[string]string
		if err := nv.NewXDRDecoder(bytes.NewReader(out)).Decode(&results); err == nil {
			snapName = results["target"]
		}
	}
	return snapName, err
}
