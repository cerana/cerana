package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	zfsp "github.com/cerana/cerana/providers/zfs"
)

func (s *zfs) TestExists() {
	tests := []struct {
		args   *zfsp.CommonArgs
		exists bool
		err    string
	}{
		{&zfsp.CommonArgs{Name: ""}, false, "missing arg: name"},
		{&zfsp.CommonArgs{Name: "ds_no_exist"}, false, ""},
		{&zfsp.CommonArgs{Name: "fs"}, true, ""},
		{&zfsp.CommonArgs{Name: "fs/1snap@snap"}, true, ""},
		{&zfsp.CommonArgs{Name: "vol/1snap"}, true, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req := acomm.NewRequest("zfs-exists")
		req.ResponseHook = s.responseHook
		s.Require().NoError(req.SetArgs(test.args), argsS)

		res, streamURL, err := s.zfs.Exists(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
			if !s.NotNil(res, argsS) {
				continue
			}

			result, ok := res.(*zfsp.ExistsResult)
			if !s.True(ok) {
				continue
			}
			s.Equal(test.exists, result.Exists, argsS)
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}
