package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestSnapshot() {
	tests := []struct {
		args *zfsp.SnapshotArgs
		err  string
	}{
		{&zfsp.SnapshotArgs{Name: "", SnapName: "Snap"}, "missing arg: name"},
		{&zfsp.SnapshotArgs{Name: "ds_no_exist"}, "missing arg: snapname"},
		{&zfsp.SnapshotArgs{Name: "fs/1snap", SnapName: "snap"}, eexist},
		{&zfsp.SnapshotArgs{Name: "fs/1snap@snap", SnapName: "snap"}, einval},
		{&zfsp.SnapshotArgs{Name: "fs/unmounted", SnapName: "snap"}, ""},
		{&zfsp.SnapshotArgs{Name: "fs/unmounted_children", SnapName: "snap", Recursive: true}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-snapshot", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Snapshot(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}
