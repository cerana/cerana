package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	zfsp "github.com/cerana/cerana/providers/zfs"
)

func (s *zfs) TestMount() {
	tests := []struct {
		args *zfsp.MountArgs
		err  string
	}{
		{&zfsp.MountArgs{Name: ""}, "missing arg: name"},
		{&zfsp.MountArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.MountArgs{Name: "fs/1snap"}, ebusy},
		{&zfsp.MountArgs{Name: "fs/unmounted"}, ""},
		{&zfsp.MountArgs{Name: "fs/unmounted_children"}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)
		req := acomm.NewRequest("zfs-mount")
		req.ResponseHook = s.responseHook
		s.Require().NoError(req.SetArgs(test.args), argsS)

		res, streamURL, err := s.zfs.Mount(req)
		s.Empty(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
