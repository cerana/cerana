package zfs_test

import (
	"fmt"
	"strings"

	"github.com/cerana/cerana/acomm"
	zfsp "github.com/cerana/cerana/providers/zfs"
	libzfs "github.com/cerana/cerana/zfs"
)

func (s *zfs) TestList() {
	typeCounts, err := expectedTypeCounts(s.pool)
	s.Require().NoError(err)

	tests := []struct {
		args *zfsp.ListArgs
		err  string
	}{
		{&zfsp.ListArgs{Name: longName}, einval},
		{&zfsp.ListArgs{Name: "ds_no_exist"}, enoent},
		{&zfsp.ListArgs{Name: s.pool}, ""},
		{&zfsp.ListArgs{Name: s.pool, Types: []string{libzfs.DatasetFilesystem}}, ""},
		{&zfsp.ListArgs{Name: s.pool, Types: []string{libzfs.DatasetVolume}}, ""},
		{&zfsp.ListArgs{Name: s.pool, Types: []string{libzfs.DatasetSnapshot}}, ""},
		{&zfsp.ListArgs{Name: s.pool, Types: []string{libzfs.DatasetVolume, libzfs.DatasetSnapshot}}, ""},
		{&zfsp.ListArgs{Name: s.pool, Types: []string{"foobar"}}, ""},
	}

	for _, test := range tests {
		argsS := fmt.Sprintf("%+v", test.args)

		req, err := acomm.NewRequest(&acomm.RequestOptions{
			Task:         "zfs-list",
			ResponseHook: s.responseHook,
			Args:         test.args,
		})
		s.Require().NoError(err, argsS)

		res, streamURL, err := s.zfs.List(req)
		s.Empty(streamURL, argsS)
		if test.err == "" {
			s.NoError(err, argsS)
			if !s.NotNil(res, argsS) {
				continue
			}

			result, ok := res.(*zfsp.ListResult)
			if !s.True(ok) {
				continue
			}
			if !s.NotNil(result.Datasets) {
				continue
			}

			expectedLen := 0
			if len(test.args.Types) == 0 {
				for _, count := range typeCounts {
					expectedLen = expectedLen + count
				}
			} else {
				for _, typeName := range test.args.Types {
					expectedLen = expectedLen + typeCounts[typeName]
				}
			}
			s.Len(result.Datasets, expectedLen, argsS)

			if len(test.args.Types) > 0 {
				for _, ds := range result.Datasets {
					s.Contains(test.args.Types, ds.Properties.Type, argsS)
				}
			}
		} else {
			s.Nil(res, argsS)
			s.EqualError(err, test.err, argsS)
		}
	}
}

func expectedTypeCounts(name string) (map[string]int, error) {
	out, err := command("sudo", "zfs", "list", "-r", "-o", "type", "-t", "all", name).Output()
	if err != nil {
		return nil, err
	}

	typeCounts := make(map[string]int)
	for _, tn := range strings.Split(string(out), "\n") {
		typeName := strings.TrimSpace(tn)
		if typeName == "" || typeName == "TYPE" {
			continue
		}
		typeCounts[typeName]++
	}
	return typeCounts, nil
}
