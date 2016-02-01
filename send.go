package main

import (
	"bytes"

	"github.com/mistifyio/gozfs/nv"
)

func send(name string, outputFD uintptr, fromSnap string, largeBlockOK, embedOK bool) error {
	innvl := map[string]interface{}{
		"fd": int32(outputFD),
	}

	// The following properties are optional, but must only be present in innvl
	// if being used.
	if fromSnap != "" {
		innvl["fromsnap"] = fromSnap
	}
	if largeBlockOK {
		innvl["largeblockok"] = true
	}
	if embedOK {
		innvl["embedok"] = true
	}

	m := map[string]interface{}{
		"cmd":     "zfs_send",
		"version": uint64(0),
		"innvl":   innvl,
	}

	encoded := &bytes.Buffer{}
	err := nv.NewNativeEncoder(encoded).Encode(m)
	if err != nil {
		return err
	}

	return ioctl(zfs, name, encoded.Bytes(), nil)
}
