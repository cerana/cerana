package gozfs

import "os"

var zfs *os.File

func init() {
	z, err := os.OpenFile("/dev/zfs", os.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	zfs = z
}
