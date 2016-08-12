package zfs

// #cgo CFLAGS: -fms-extensions -Wno-microsoft
// int zfs_ioctl(int, char *, int, void *, int, void *, int);
import "C"
import (
	"os"
	"unsafe"

	"github.com/cerana/cerana/pkg/errors"
)

func ioctl(f *os.File, name string, input, output []byte) error {
	if len(input) == 0 {
		return errors.New("input nvl required")
	}
	in := unsafe.Pointer(&input[0])

	var out unsafe.Pointer
	if len(output) != 0 {
		out = unsafe.Pointer(&output[0])
	}

	_, err := C.zfs_ioctl(C.int(f.Fd()),
		C.CString(name), C.int(len(name)),
		unsafe.Pointer(in), C.int(len(input)),
		unsafe.Pointer(out), C.int(len(output)))
	return errors.Wrapv(err, map[string]interface{}{"name": name, "args": input})
}
