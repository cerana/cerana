package main

import (
	"net"
	"sort"

	"github.com/cerana/cerana/providers/clusterconf"
	zfsp "github.com/cerana/cerana/providers/zfs"
	"github.com/cerana/cerana/zfs"
)

func (s *StatsPusher) TestGetDatasets() {
	tests := []struct {
		desc     string
		datasets []string
		result   []string
	}{
		{"empty", []string{}, []string{}},
		{"one", []string{"asdf"}, []string{"asdf"}},
		{"base stripping", []string{"base/asdf"}, []string{"asdf"}},
		{"base stripping", []string{"foobar", "foobar/asdf"}, []string{"asdf"}},
	}

	for _, test := range tests {
		s.zfs.Data.Datasets = make(map[string]*zfsp.Dataset)
		for _, name := range test.datasets {
			s.zfs.Data.Datasets[name] = &zfsp.Dataset{Name: name, Properties: &zfs.DatasetProperties{Type: "volume"}}
		}
		datasets, err := s.statsPusher.getDatasets()
		if !s.NoError(err, test.desc) {
			continue
		}
		sort.Strings(test.result)
		sort.Strings(datasets)
		s.Equal(test.result, datasets, test.desc)
	}
}

func (s *StatsPusher) TestGetIP() {
	ip, err := s.statsPusher.getIP()
	if !s.NoError(err) {
		return
	}
	s.EqualValues(s.metrics.Data.Network.Interfaces[0].Addrs[0].Addr, ip.String())
}

func (s *StatsPusher) TestSendDatasetHeartbeats() {
	name := "foobar"
	s.zfs.Data.Datasets = make(map[string]*zfsp.Dataset)
	s.zfs.Data.Datasets[name] = &zfsp.Dataset{Name: name, Properties: &zfs.DatasetProperties{Type: "volume"}}
	s.clusterConf.Data.Datasets = make(map[string]*clusterconf.Dataset)
	s.clusterConf.Data.Datasets[name] = &clusterconf.Dataset{DatasetConf: clusterconf.DatasetConf{ID: name}}
	ip := net.ParseIP(s.metrics.Data.Network.Interfaces[0].Addrs[0].Addr)
	s.NoError(s.statsPusher.sendDatasetHeartbeats([]string{name}, ip))
	s.True(s.clusterConf.Data.Datasets[name].Nodes[ip.String()])
}
