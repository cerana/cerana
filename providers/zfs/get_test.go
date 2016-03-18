package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestGet() {
	tests := []struct {
		args *zfsp.CommonArgs
		err  string
	}{
		{&zfsp.CommonArgs{Name: ""}, "missing arg: name"},
		{&zfsp.CommonArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.CommonArgs{Name: "fs"}, ""},
		{&zfsp.CommonArgs{Name: "fs/1snap@snap"}, ""},
		{&zfsp.CommonArgs{Name: "vol/1snap"}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-get", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Get(req)
		s.Empty(streamURL, argsS)
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
