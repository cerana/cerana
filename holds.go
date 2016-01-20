package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

const emptyList = "\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00"

func holds(name string) ([]string, error) {
	m := map[string]interface{}{
		"cmd":     "zfs_get_holds",
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err := nv.Encode(encoded, m)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 1024)
	copy(out, emptyList)

	err = ioctl(zfs, name, encoded.Bytes(), out)
	if err != nil {
		return nil, err
	}

	m = map[string]interface{}{}

	if err = nv.NewXDRDecoder(bytes.NewReader(out)).Decode(&m); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	return names, nil
}
