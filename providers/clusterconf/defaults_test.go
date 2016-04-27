package clusterconf_test

import (
	"encoding/json"
	"path"
	"strconv"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/mistifyio/lochness/pkg/kv"
)

func (s *clusterConf) TestGetDefaults() {
	tests := []struct {
		zfsManual bool
	}{
		{true},
		{false},
	}

	for _, test := range tests {
		desc := strconv.FormatBool(test.zfsManual)
		_ = s.setDefaultZFSManual(test.zfsManual)
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "get-defaults",
		})
		s.Require().NoError(err, desc)
		result, streamURL, err := s.clusterConf.GetDefaults(req)
		s.Nil(streamURL, desc)
		s.NoError(err, desc)
		if !s.NotNil(result, desc) {
			continue
		}
		defaultsPayload, ok := result.(*clusterconf.DefaultsPayload)
		s.True(ok, desc)
		s.Equal(test.zfsManual, defaultsPayload.Defaults.ZFSManual)
	}
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
					DefaultsConf: &clusterconf.DefaultsConf{ZFSManual: test.zfsManual},
					ModIndex:     *test.modIndex,
				},
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.UpdateDefaults(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.EqualError(err, test.err, test.desc)
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

func (s *clusterConf) setDefaultZFSManual(value bool) *clusterconf.Defaults {
	defaults := &clusterconf.Defaults{DefaultsConf: &clusterconf.DefaultsConf{ZFSManual: value}}
	sj, _ := json.Marshal(defaults)
	key := path.Join("cluster")
	s.kvp.Data[key] = kv.Value{Data: sj, Index: 1}
	defaults.ModIndex = 1
	return defaults
}
