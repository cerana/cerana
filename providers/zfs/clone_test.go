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
		{&zfsp.CloneArgs{Name: "", Origin: "fs/1snap@snap"}, "missing arg: name"},
		{&zfsp.CloneArgs{Name: "missing_origin", Origin: ""}, "missing arg: origin"},

		{&zfsp.CloneArgs{Name: "fs/1snap", Origin: "fs/1snap@snap"}, eexist},
		{&zfsp.CloneArgs{Name: "fs/no_exist_origin", Origin: "fs/1snap@foobar"}, enoent},
		{&zfsp.CloneArgs{Name: "fs/basic_clone", Origin: "fs/1snap@snap"}, ""},
		{&zfsp.CloneArgs{Name: "fs/clone_with_props", Origin: "fs/1snap@snap", Properties: props{"foo:bar": "baz"}}, ""},

		{&zfsp.CloneArgs{Name: "vol/1snap", Origin: "vol/1snap@snap"}, eexist},
		{&zfsp.CloneArgs{Name: "vol/no_exist_origin", Origin: "vol/1snap@foobar"}, enoent},
		{&zfsp.CloneArgs{Name: "vol/basic_clone", Origin: "vol/1snap@snap"}, ""},
		{&zfsp.CloneArgs{Name: "vol/clone_with_props", Origin: "vol/1snap@snap", Properties: props{"foo:bar": "baz"}}, ""},
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
