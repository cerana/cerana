package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/mistifyio/gozfs/nv"
)

type header struct {
	Size     uint32
	ExtSpace uint8
	Error    uint8
	Endian   uint8
	Reserved uint8
}

func getSize(b []byte) (int64, error) {
	h := header{}
	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.LittleEndian, &h)
	if err != nil {
		return 0, err
	}

	if h.Endian != 1 {
		buf := bytes.NewBuffer(b)
		err := binary.Read(buf, binary.BigEndian, &h)
		if err != nil {
			return 0, err
		}
	}

	if h.Reserved != 0 {
		return 0, errors.New("non-zero Reserved field")
	}
	if h.Endian > 1 {
		return 0, errors.New("unknown Endian value")
	}
	if h.Error != 0 {
		return 0, syscall.Errno(h.Error)
	}

	return int64(h.Size), nil
}

func properties(name string, types map[string]bool, recurse bool, depth uint64) (map[string]interface{}, error) {
	listing, err := list(name, types, recurse, depth)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]interface{}, len(listing))
	for _, l := range listing {
		name := l["name"].(string)
		props := l["properties"].(map[string]interface{})
		ret[name] = props
	}
	return ret, nil
}

func list(name string, types map[string]bool, recurse bool, depth uint64) ([]map[string]interface{}, error) {
	var reader io.Reader
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer reader.(*os.File).Close()
	defer writer.Close()

	opts := map[string]interface{}{
		"fd": int32(writer.Fd()),
	}
	if types != nil {
		opts["type"] = types
	}
	if recurse != false {
		if depth != 0 {
			opts["recurse"] = depth
		} else {
			opts["recurse"] = true
		}
	}
	args := map[string]interface{}{
		"cmd":     "zfs_list",
		"innvl":   map[string]interface{}{},
		"opts":    opts,
		"version": uint64(0),
	}

	encoded := &bytes.Buffer{}
	err = nv.Encode(encoded, args)
	if err != nil {
		return nil, err
	}

	err = ioctl(zfs, name, encoded.Bytes(), nil)
	if err != nil {
		return nil, err
	}

	var buf []byte
	reader = bufio.NewReader(reader)

	ret := []map[string]interface{}{}
	for {
		header := make([]byte, 8)
		_, err = io.ReadFull(reader, header)
		if err != nil {
			break
		}

		var size int64
		size, err = getSize(header)
		if err != nil {
			break
		}
		if size == 0 {
			break
		}

		if len(buf) < int(size) {
			l := (size + 1023) & ^1023
			buf = make([]byte, l)
		}
		buf = buf[:size]

		_, err = io.ReadFull(reader, buf)
		if err != nil {
			break
		}

		m := map[string]interface{}{}
		err = nv.NewXDRDecoder(bytes.NewReader(buf)).Decode(&m)
		if err != nil {
			break
		}
		ret = append(ret, m)
	}
	return ret, err
}
