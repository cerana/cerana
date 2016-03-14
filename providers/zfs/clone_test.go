package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestClone() {
	tests := []struct {
		args *zfsp.CloneArgs
		err  string
	}{
		{&zfsp.CloneArgs{Name: "", Origin: "fs1/fs1@snap1"}, "missing arg: name"},
		{&zfsp.CloneArgs{Name: "fs1/clone1", Origin: ""}, "missing arg: origin"},
		{&zfsp.CloneArgs{Name: "fs1/fs2", Origin: "fs1/fs1@snap1"}, eexist},
		{&zfsp.CloneArgs{Name: "fs1/clone2", Origin: "fs1/fs1@foobar"}, enoent},
		{&zfsp.CloneArgs{Name: "fs1/clone4", Origin: "fs1/fs1@snap1"}, ""},
		{&zfsp.CloneArgs{Name: "fs1/clone5", Origin: "fs1/fs1@snap1", Properties: map[string]interface{}{"foo:bar": "baz"}}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		if test.args.Origin != "" {
			test.args.Origin = filepath.Join(s.pool, test.args.Origin)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-clone", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Clone(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			if !s.Nil(err, argsS) {
				continue
			}
			if !s.NotNil(res, argsS) {
				continue
			}

			result, ok := res.(*zfsp.DatasetResult)
			if !s.True(ok, argsS) {
				continue
			}
			if !s.NotNil(result.Dataset) {
				continue
			}
			s.Equal(result.Dataset.Name, test.args.Name, argsS)
			s.Equal(result.Dataset.Properties.Origin, test.args.Origin, argsS)
			for key, value := range test.args.Properties {
				s.Equal(result.Dataset.Properties.UserDefined[key], value, argsS)
			}
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
