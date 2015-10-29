package main

import "github.com/mistifyio/gozfs/nv"

func rollback(name string) (string, error) {
	m := map[string]interface{}{
		"cmd":     "zfs_rollback",
		"version": uint64(0),
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return "", err
	}

	var snapName string
	out := make([]byte, 1024)
	err = ioctl(zfs, name, encoded, out)
	if err == nil {
		var results map[string]string
		if err := nv.Decode(out, &results); err == nil {
			snapName = results["target"]
		}
	}
	return snapName, err
}
