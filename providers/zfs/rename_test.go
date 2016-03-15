package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestRename() {
	tests := []struct {
		args *zfsp.RenameArgs
		err  string
	}{
		{&zfsp.RenameArgs{Name: "", Origin: "fs/1snap"}, "missing arg: name"},
		{&zfsp.RenameArgs{Name: "new_no_origin", Origin: ""}, "missing arg: origin"},
		{&zfsp.RenameArgs{Name: "new_no_exist", Origin: "ds_no_exist"}, enoent},
		{&zfsp.RenameArgs{Name: "fs/1snap@snap_new", Origin: "fs/1snap@snap"}, ""},
		{&zfsp.RenameArgs{Name: "vol/1snap_new", Origin: "vol/1snap"}, ""},
		{&zfsp.RenameArgs{Name: "fs_new", Origin: "fs"}, ""},
	}

	for _, test := range tests {
		if test.args.Origin != "" {
			test.args.Origin = filepath.Join(s.pool, test.args.Origin)
		}
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-rename", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Rename(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}
