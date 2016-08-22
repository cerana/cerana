package zfs

import (
	"os"
	"sync"

	"github.com/cerana/cerana/pkg/errors"
)

// zfsFD is the zfs device file descriptor. Do not use this directly. Access by
// calling zfs() instead. The initialization is protected by once.
var zfsFD *os.File
var once sync.Once

func openZFS() {
	z, err := os.OpenFile("/dev/zfs", os.O_RDWR, 0)
	if err != nil {
		panic(errors.Wrap(err))
	}
	zfsFD = z
}

func zfs() *os.File {
	once.Do(openZFS)
	return zfsFD
}
