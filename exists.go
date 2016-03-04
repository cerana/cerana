package gozfs

import (
	"bytes"
	"syscall"

	"github.com/mistifyio/gozfs/nv"
)

func exists(name string) error {
	m := map[string]interface{}{
		"cmd":     "zfs_exists",
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded.Bytes(), nil)
}

// Exists determines whether a dataset exists or not.
func Exists(name string) (bool, error) {
	if err := exists(name); err != nil {
		if err.Error() == syscall.ENOENT.Error() {
			err = nil
		}
		return false, err
	}
	return true, nil
}
