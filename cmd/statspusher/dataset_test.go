package main

import (
	"net"

	"github.com/cerana/cerana/providers/clusterconf"
	zfsp "github.com/cerana/cerana/providers/zfs"
	"github.com/cerana/cerana/zfs"
)

func (s *StatsPusher) TestGetDatasets() {
	ip := net.ParseIP("192.168.1.1")
	tests := []struct {
		desc     string
		datasets []string
		result   []clusterconf.DatasetHeartbeatArgs
	}{
		{"empty", []string{}, []clusterconf.DatasetHeartbeatArgs{}},
		{"one", []string{"asdf"}, []clusterconf.DatasetHeartbeatArgs{{ID: "asdf"}}},
		{"base stripping", []string{"base/asdf"}, []clusterconf.DatasetHeartbeatArgs{{ID: "asdf"}}},
		{"base stripping", []string{"foobar", "foobar/asdf"}, []clusterconf.DatasetHeartbeatArgs{{ID: "asdf"}}},
		{"in use", []string{"useddataset"}, []clusterconf.DatasetHeartbeatArgs{{ID: "useddataset", InUse: true}}},
	}

	for _, test := range tests {
		s.zfs.Data.Datasets = make(map[string]*zfsp.Dataset)
		for _, name := range test.datasets {
			s.zfs.Data.Datasets[name] = &zfsp.Dataset{Name: name, Properties: &zfs.DatasetProperties{Type: "volume"}}
			if test.desc == "in use" {
				bundleID := uint64(1)
				s.clusterConf.Data.Bundles[bundleID] = &clusterconf.Bundle{
					ID: bundleID,
					Datasets: map[string]clusterconf.BundleDataset{
						test.datasets[0]: {
							Name: test.datasets[0],
							ID:   test.datasets[0],
						},
					},
				}
				s.clusterConf.Data.BundlesHB[bundleID] = clusterconf.BundleHeartbeats{
					"someserial": {IP: ip},
				}
			}
		}
		datasets, err := s.statsPusher.getDatasets(ip)
		if !s.NoError(err, test.desc) {
			continue
		}

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
	s.clusterConf.Data.Datasets[name] = &clusterconf.Dataset{ID: name}
	ip := net.ParseIP(s.metrics.Data.Network.Interfaces[0].Addrs[0].Addr)
	s.NoError(s.statsPusher.sendDatasetHeartbeats([]clusterconf.DatasetHeartbeatArgs{{ID: name}}, ip))
}
