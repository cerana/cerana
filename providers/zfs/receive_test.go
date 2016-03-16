package zfs_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestReceive() {
	tests := []struct {
		args   *zfsp.CommonArgs
		origin string
		err    string
	}{
		{&zfsp.CommonArgs{Name: ""}, "fs/1snap@snap", "missing arg: name"},
		{&zfsp.CommonArgs{Name: "fs"}, "", "missing request stream-url"},
		{&zfsp.CommonArgs{Name: "does_not_exist/1snap_snap_received"}, "fs/1snap@snap", enoent},
		{&zfsp.CommonArgs{Name: "fs/unmounted_received"}, "fs/unmounted", ""},
		{&zfsp.CommonArgs{Name: "fs/1snap_snap_received"}, "fs/1snap@snap", ""},
		{&zfsp.CommonArgs{Name: "fs/1snap"}, "fs/1snap@snap", eexist},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		if test.origin != "" {
			test.origin = filepath.Join(s.pool, test.origin)
		}
		argsS := fmt.Sprintf("%+v, origin: %s", test.args, test.origin)

		reqStreamURL := ""
		if test.origin != "" {
			streamURL, err := s.zfsSendStreamURL(test.origin)
			s.Require().NoError(err)
			reqStreamURL = streamURL.String()
		}

		req, err := acomm.NewRequest("zfs-send", "unix:///tmp/foobar", reqStreamURL, test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, resStreamURL, err := s.zfs.Receive(req)
		s.Nil(resStreamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
			if !s.NotNil(res, argsS) {
				continue
			}

			result, ok := res.(*zfsp.DatasetResult)
			if !s.True(ok) {
				continue
			}
			if !s.NotNil(result.Dataset) {
				continue
			}
			s.Equal(test.args.Name, result.Dataset.Name, argsS)
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}

func (s *zfs) zfsSendStreamURL(name string) (*url.URL, error) {
	cmd := command("sudo", "zfs", "send", name)
	var stream bytes.Buffer
	cmd.Stdout = &stream
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return s.tracker.NewStreamUnix(s.config.StreamDir("zfs-receive"), ioutil.NopCloser(&stream))
}
