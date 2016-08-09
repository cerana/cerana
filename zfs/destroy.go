package zfs

import (
	"bytes"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs/nv"
)

func destroy(name string, deferFlag bool) error {
	m := map[string]interface{}{
		"cmd":     "zfs_destroy",
		"version": uint64(0),
		"defer":   deferFlag,
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"name": name, "input": m})
	}

	return ioctl(zfs(), name, encoded.Bytes(), nil)
}
