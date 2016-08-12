package zfs

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"os"
	"syscall"

	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/errorutils"
	"github.com/cerana/cerana/zfs/nv"
)

type pipeHeader struct {
	Size     uint32
	ExtSpace uint8
	Error    uint8
	Endian   uint8
	Reserved uint8
}

func getSize(r io.Reader) (int64, error) {
	buf := make([]byte, 8)

	_, err := io.ReadFull(r, buf)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read size data")
	}

	h := pipeHeader{}
	err = binary.Read(bytes.NewReader(buf), binary.LittleEndian, &h)
	if err != nil {
		return 0, errors.Wrap(err, "failed to read little endian pipe header")
	}

	if h.Endian != 1 {
		err := binary.Read(bytes.NewReader(buf), binary.BigEndian, &h)
		if err != nil {
			return 0, errors.Wrap(err, "failed to read big endian pipe header")
		}
	}

	if h.Reserved != 0 {
		return 0, errors.New("non-zero Reserved field")
	}
	if h.Endian > 1 {
		return 0, errors.New("unknown Endian value")
	}
	if h.Error != 0 {
		return 0, errors.Wrap(syscall.Errno(h.Error))
	}

	return int64(h.Size), nil
}

func properties(name string, types map[string]bool, recurse bool, depth uint64) (map[string]*datasetProperties, error) {
	listing, err := list(name, types, recurse, depth)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]*datasetProperties, len(listing))
	for _, l := range listing {
		name := l.Name
		props := l.Properties
		ret[name] = props
	}
	return ret, nil
}

func list(name string, types map[string]bool, recurse bool, depth uint64) (ret []*dataset, err error) {
	pipeReader, writer, err := os.Pipe()
	if err != nil {
		return
	}

	var errData map[string]interface{}
	defer func() {
		err = errorutils.First(err, errors.Wrapv(writer.Close(), errData), errors.Wrapv(pipeReader.Close(), errData))
	}()

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

	errData = map[string]interface{}{"name": name, "args": args}

	encoded := &bytes.Buffer{}
	err = nv.NewNativeEncoder(encoded).Encode(args)
	if err != nil {
		err = errors.Wrapv(err, errData)
		return
	}

	err = ioctl(zfs(), name, encoded.Bytes(), nil)
	if err != nil {
		return
	}

	var buf []byte
	reader := bufio.NewReader(pipeReader)

	for {
		var size int64
		size, err = getSize(reader)
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
			err = errors.Wrapv(err, errData)
			break
		}

		m := &dataset{}
		err = nv.NewXDRDecoder(bytes.NewReader(buf)).Decode(&m)
		if err != nil {
			err = errors.Wrapv(err, errData)
			break
		}

		if m.Properties.Clones.Value == nil {
			m.Properties.Clones.Value = make(map[string]nv.Boolean)
		}

		ret = append(ret, m)
	}
	return
}
