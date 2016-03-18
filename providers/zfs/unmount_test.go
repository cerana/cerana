package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestUnmount() {
	tests := []struct {
		args *zfsp.UnmountArgs
		err  string
	}{
		{&zfsp.UnmountArgs{Name: ""}, "missing arg: name"},
		{&zfsp.UnmountArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.UnmountArgs{Name: "fs/unmounted"}, einval},
		{&zfsp.UnmountArgs{Name: "fs/1snap"}, ""},
		{&zfsp.UnmountArgs{Name: "fs"}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-unmount", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Unmount(req)
		s.Empty(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
