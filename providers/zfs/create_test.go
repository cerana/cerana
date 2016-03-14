package zfs_test

import (
	"fmt"
	"path/filepath"

	"github.com/mistifyio/gozfs"
	"github.com/mistifyio/mistify/acomm"
	zfsp "github.com/mistifyio/mistify/providers/zfs"
)

func (s *zfs) TestCreate() {
	fs := gozfs.DatasetFilesystem
	vol := gozfs.DatasetVolume
	tests := []struct {
		args *zfsp.CreateArgs
		err  string
	}{
		{&zfsp.CreateArgs{Name: "", Type: fs}, "missing arg: name"},
		{&zfsp.CreateArgs{Name: "fs/no_type", Type: "asdf"}, "missing or invalid arg: type"},

		{&zfsp.CreateArgs{Name: "fs", Type: fs}, eexist},
		{&zfsp.CreateArgs{Name: "fs/~1", Type: fs}, einval},
		{&zfsp.CreateArgs{Name: "fs/bad_prop", Type: fs, Properties: props{"foo": "bar"}}, einval},
		{&zfsp.CreateArgs{Name: "fs_no_exist/fs", Type: fs}, enoent},
		{&zfsp.CreateArgs{Name: "fs/basic_fs", Type: fs}, ""},
		{&zfsp.CreateArgs{Name: "fs/fs_with_props", Type: fs, Properties: props{"foo:bar": "baz"}}, ""},

		{&zfsp.CreateArgs{Name: "", Type: vol, Properties: props{"volsize": 8192}}, "missing arg: name"},
		{&zfsp.CreateArgs{Name: "vol/no_size", Type: vol, Properties: nil}, "missing or invalid arg: volsize"},
		{&zfsp.CreateArgs{Name: "vol/bad_size", Type: vol, Volsize: 0, Properties: nil}, "missing or invalid arg: volsize"},
		{&zfsp.CreateArgs{Name: "vol/bad_prop", Type: vol, Volsize: 8192, Properties: props{"foo": "bar"}}, einval},
		{&zfsp.CreateArgs{Name: "vol/1snap", Type: vol, Volsize: 8192, Properties: nil}, eexist},
		{&zfsp.CreateArgs{Name: "vol/basic_vol", Type: vol, Volsize: 8192, Properties: nil}, ""},
		{&zfsp.CreateArgs{Name: "vol/vol_with_blocksize", Type: vol, Volsize: 1024, Properties: props{"volblocksize": 1024}}, ""},
	}

	for _, test := range tests {
		if test.args.Name != "" {
			test.args.Name = filepath.Join(s.pool, test.args.Name)
		}
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest("zfs-create", "unix:///tmp/foobar", "", test.args, nil, nil)
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.Create(req)
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
			ds := result.Dataset
			if !s.NotNil(ds, argsS) {
				continue
			}
			s.Equal(ds.Name, test.args.Name, argsS)
			s.Equal(ds.Properties.Type, test.args.Type, argsS)
			if _, ok := test.args.Properties["foo:bar"]; ok {
				s.Equal(ds.Properties.UserDefined["foo:bar"], "baz", argsS)
			}
			if test.args.Type == vol {

			}
		} else {
			s.EqualError(err, test.err, argsS)
		}
	}
}
