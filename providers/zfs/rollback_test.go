package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	zfsp "github.com/cerana/cerana/providers/zfs"
)

func (s *zfs) TestRollback() {
	tests := []struct {
		args *zfsp.RollbackArgs
		err  string
	}{
		{&zfsp.RollbackArgs{Name: ""}, "missing arg: name"},
		{&zfsp.RollbackArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.RollbackArgs{Name: "fs/1snap"}, ""},
		{&zfsp.RollbackArgs{Name: "fs/1snap@snap"}, ""},
		{&zfsp.RollbackArgs{Name: "fs/3snap@snap2"}, "not most recent snapshot"},
		{&zfsp.RollbackArgs{Name: "fs/3snap@snap1", DestroyRecent: true}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest(&acomm.RequestOptions{
			Task:         "zfs-rollback",
			ResponseHook: s.responseHook,
			Args:         test.args,
		})
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Rollback(req)
		s.Empty(streamURL, argsS)
		s.Nil(res, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
