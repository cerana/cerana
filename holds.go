package main

import "github.com/mistifyio/gozfs/nv"

func holds(name string) ([]string, error) {
	m := map[string]interface{}{
		"cmd":     "zfs_get_holds",
		"version": uint64(0),
	}

	encoded, err := nv.Encode(m)
	if err != nil {
		return nil, err
	}

	out := make([]byte, 1024)
	err = ioctl(zfs, name, encoded, out)
	if err != nil {
		return nil, err
	}

	m = map[string]interface{}{}

	if err = nv.Decode(out, &m); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}

	return names, nil
}
