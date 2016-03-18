package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestDestroy() {
	tests := []struct {
		args *zfsp.DestroyArgs
		err  string
	}{
		{&zfsp.DestroyArgs{Name: ""}, "missing arg: name"},
		{&zfsp.DestroyArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.DestroyArgs{Name: "fs/1snap"}, ebusy},
		{&zfsp.DestroyArgs{Name: "fs/hold_snap@snap"}, ebusy},
		{&zfsp.DestroyArgs{Name: "fs/unmounted"}, ""},
		{&zfsp.DestroyArgs{Name: "fs/unmounted_children"}, eexist},
		{&zfsp.DestroyArgs{Name: "fs/unmounted_children", Recursive: true}, ""},
		{&zfsp.DestroyArgs{Name: "fs@snap_with_clone"}, eexist},
		{&zfsp.DestroyArgs{Name: "fs@snap_with_clone", Recursive: true}, eexist},
		{&zfsp.DestroyArgs{Name: "fs@snap_with_clone", RecursiveClones: true}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-destroy", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Destroy(req)
		s.Empty(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
