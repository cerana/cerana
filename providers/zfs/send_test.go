package zfs_test

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestSend() {
	tests := []struct {
		args *zfsp.CommonArgs
		err  string
	}{
		{&zfsp.CommonArgs{Name: ""}, "missing arg: name"},
		{&zfsp.CommonArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.CommonArgs{Name: "fs/1snap@snap"}, ""},
		{&zfsp.CommonArgs{Name: "fs"}, ""},
		{&zfsp.CommonArgs{Name: "vol/1snap"}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-send", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Send(req)
		if test.err == "" {
			s.NoError(err, argsS)
			if !s.NotEmpty(streamURL, argsS) {
				continue
			}

			var stream bytes.Buffer
			if !s.NoError(acomm.Stream(&stream, streamURL), argsS) {
				continue
			}
			s.True(verifyZFSStream(&stream), argsS)
		} else {
			s.Nil(res, argsS)
			s.Empty(streamURL, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}

func verifyZFSStream(stream io.Reader) bool {
	cmd := command("zstreamdump")
	cmd.Stdin = stream
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
