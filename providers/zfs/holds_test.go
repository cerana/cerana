package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestHolds() {
	tests := []struct {
		args  *zfsp.CommonArgs
		holds []string
		err   string
	}{
		{&zfsp.CommonArgs{Name: ""}, nil, "missing arg: name"},
		{&zfsp.CommonArgs{Name: "ds_no_exist"}, nil, enoent},
		{&zfsp.CommonArgs{Name: "fs"}, nil, ""},
		{&zfsp.CommonArgs{Name: "fs/hold_snap@snap"}, []string{"hold"}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-holds", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Holds(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
			if !s.NotNil(res, argsS) {
				continue
			}

			result, ok := res.(*zfsp.HoldsResult)
			if !s.True(ok) {
				continue
			}
			s.Len(result.Holds, len(test.holds), argsS)
			for _, hn := range test.holds {
				s.Contains(result.Holds, hn, argsS)
			}
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}
