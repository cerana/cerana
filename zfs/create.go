package gozfs

import (
	"bytes"

	"github.com/cerana/cerana/zfs/nv"
)

func create(name string, createType dmuType, props map[string]interface{}) error {
	m := map[string]interface{}{
		"cmd":     "zfs_create",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"type":  createType,
			"props": props,
		},
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded.Bytes(), nil)
}
