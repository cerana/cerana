package zfs

import (
	"bytes"

	"github.com/cerana/cerana/zfs/nv"
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

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return "", err
	}

	out := make([]byte, 1024)
	err = ioctl(zfs(), name, encoded.Bytes(), out)

	var failedName string
	if err != nil && recursive {
		_ = nv.NewNativeDecoder(bytes.NewReader(out)).Decode(&failedName)
	}
	return failedName, err
}
