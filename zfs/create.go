package zfs

import (
	"bytes"

	"github.com/cerana/cerana/pkg/errors"
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
		return errors.Wrapv(err, map[string]interface{}{"name": name, "args": m})
	}

	return ioctl(zfs(), name, encoded.Bytes(), nil)
}
