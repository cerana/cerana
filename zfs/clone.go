package zfs

import (
	"bytes"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs/nv"
)

func clone(name, origin string, props map[string]interface{}) error {
	m := map[string]interface{}{
		"cmd":     "zfs_clone",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"origin": origin,
			"props":  props,
		},
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"name": name, "args": m})
	}

	return ioctl(zfs(), name, encoded.Bytes(), nil)
}
