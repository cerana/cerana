package zfs

import (
	"bytes"
	"syscall"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs/nv"
)

func exists(name string) error {
	m := map[string]interface{}{
		"cmd":     "zfs_exists",
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"name": name, "args": m})
	}

	return ioctl(zfs(), name, encoded.Bytes(), nil)
}

// Exists determines whether a dataset exists or not.
func Exists(name string) (bool, error) {
	if err := exists(name); err != nil {
		if errors.Cause(err) == syscall.ENOENT {
			err = nil
		}
		return false, err
	}
	return true, nil
}
