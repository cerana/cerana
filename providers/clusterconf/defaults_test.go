package clusterconf_test

import (
	"path"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
)

func (s *clusterConf) TestGetDefaults() {
	_, err := s.setDefaultZFSManual(true)
	s.Require().NoError(err)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "get-defaults",
	})
	s.Require().NoError(err)
	result, streamURL, err := s.clusterConf.GetDefaults(req)
	s.Nil(streamURL)
	s.NoError(err)
	if !s.NotNil(result) {
		return
	}
	defaultsPayload, ok := result.(*clusterconf.DefaultsPayload)
	s.True(ok)
	s.True(defaultsPayload.Defaults.ZFSManual)
}

func (s *clusterConf) TestUpdateDefaults() {
	var modIndex uint64
	tests := []struct {
		desc      string
		zfsManual bool
		modIndex  *uint64
		err       string
	}{
		{"first set", true, new(uint64), ""},
		{"bad index", false, new(uint64), "CAS failed"},
		{"good index", false, &modIndex, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "update-defaults",
			Args: &clusterconf.DefaultsPayload{
				Defaults: &clusterconf.Defaults{
					DefaultsConf: clusterconf.DefaultsConf{ZFSManual: test.zfsManual},
					ModIndex:     *test.modIndex,
				},
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.UpdateDefaults(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.Contains(err.Error(), test.err, test.desc)
			s.Nil(result, test.desc)
		} else {
			s.NoError(err, test.desc)
			if !s.NotNil(result, test.desc) {
				continue
			}
			defaultsPayload, ok := result.(*clusterconf.DefaultsPayload)
			if !s.True(ok, test.desc) {
				continue
			}
			s.Equal(test.zfsManual, defaultsPayload.Defaults.ZFSManual, test.desc)
			s.NotEqual(*test.modIndex, defaultsPayload.Defaults.ModIndex, test.desc)
			if defaultsPayload.Defaults.ModIndex != 0 {
				modIndex = defaultsPayload.Defaults.ModIndex
			}
		}
	}
}

func (s *clusterConf) setDefaultZFSManual(value bool) (*clusterconf.Defaults, error) {
	defaults := &clusterconf.Defaults{DefaultsConf: clusterconf.DefaultsConf{ZFSManual: value}}
	key := path.Join("cluster")
	data := map[string]interface{}{
		key: defaults,
	}
	indexes, err := s.loadData(data)
	if err != nil {
		return nil, err
	}
	defaults.ModIndex = indexes[key]
	return defaults, nil
}
