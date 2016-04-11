package zfs

import (
	"io"
	"strings"
	"syscall"

	gzfs "github.com/mistifyio/go-zfs"
)

func receive(inputFD uintptr, name string) error { return nil }

// Receive creates a snapshot from a zfs send stream.
func Receive(stream io.Reader, name string) error {
	// TODO: Reimplement when we have a native zfs_receive
	if _, err := gzfs.ReceiveSnapshot(stream, name); err != nil {
		// Fix errors to be more like what zfs will probably return
		if strings.Contains(err.Error(), "dataset does not exist") {
			return syscall.ENOENT
		} else if strings.Contains(err.Error(), "exists\nmust specify -F to overwrite") {
			return syscall.EEXIST
		}
		return err
	}
	return nil
}
