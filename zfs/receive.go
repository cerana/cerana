package zfs

import (
	"io"
	"strings"
	"syscall"

	"github.com/cerana/cerana/pkg/errors"
	gzfs "github.com/mistifyio/go-zfs"
)

func receive(inputFD uintptr, name string) error { return nil }

// Receive creates a snapshot from a zfs send stream.
func Receive(stream io.Reader, name string) error {
	// TODO: Reimplement when we have a native zfs_receive
	errData := map[string]interface{}{"name": name}
	if _, err := gzfs.ReceiveSnapshot(stream, name); err != nil {
		// Fix errors to be more like what zfs will probably return
		if strings.Contains(err.Error(), "dataset does not exist") {
			return errors.Wrapv(syscall.ENOENT, errData)
		} else if strings.Contains(err.Error(), "exists\nmust specify -F to overwrite") {
			return errors.Wrapv(syscall.EEXIST, errData)
		}
		return errors.Wrapv(err, errData)
	}
	return nil
}
