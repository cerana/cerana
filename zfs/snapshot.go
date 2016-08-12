package zfs

import (
	"bytes"
	"syscall"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/zfs/nv"
)

func snapshot(pool string, snapNames []string, props map[string]string) (map[string]syscall.Errno, error) {
	// snaps needs to be a map with the snap name as the key and an arbitrary value
	snaps := make(map[string]string)
	for _, snapName := range snapNames {
		snaps[snapName] = ""
	}

	m := map[string]interface{}{
		"cmd":     "zfs_snapshot",
		"version": uint64(0),
		"innvl": map[string]interface{}{
			"snaps": snaps,
			"props": props,
		},
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return nil, errors.Wrapv(err, map[string]interface{}{"name": pool, "args": m})
	}

	out := make([]byte, 1024)
	err = ioctl(zfs(), pool, encoded.Bytes(), out)

	var errlist map[string]syscall.Errno
	if errno, ok := errors.Cause(err).(syscall.Errno); ok && errno == syscall.EEXIST {
		// Try to get errlist info, but ignore any errors in the attempt
		errs := map[string]int32{}
		if nv.NewNativeDecoder(bytes.NewReader(out)).Decode(&errs) == nil {
			errlist = make(map[string]syscall.Errno, len(errs))
			for snap, errno := range errs {
				errlist[snap] = syscall.Errno(errno)
			}
		}
	}
	return errlist, err
}
