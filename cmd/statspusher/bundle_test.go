package main

import (
	"fmt"
	"net"
	"sort"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/coreos/go-systemd/dbus"
	"github.com/pborman/uuid"
)

type uint64s []uint64

func (u uint64s) Len() int           { return len(u) }
func (u uint64s) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u uint64s) Less(i, j int) bool { return u[i] < u[j] }

func (s *StatsPusher) TestGetSerial() {
	serial, err := s.statsPusher.getSerial()
	if !s.NoError(err) {
		return
	}
	s.Equal(s.metrics.Data.Host.Hostname, serial)
}

func (s *StatsPusher) extractBundles() {

}

func (s *StatsPusher) TestGetBundles() {
	tests := []struct {
		desc   string
		known  []uint64
		local  []uint64
		result uint64s
	}{
		{"empty", []uint64{}, []uint64{}, []uint64{}},
		{"known only", []uint64{123}, []uint64{}, []uint64{}},
		{"local only", []uint64{}, []uint64{123}, []uint64{}},
		{"both, single", []uint64{123}, []uint64{123}, []uint64{123}},
		{"extra local", []uint64{123}, []uint64{123, 456}, []uint64{123}},
		{"extra known", []uint64{123, 456}, []uint64{123}, []uint64{123}},
		{"both, multiple", []uint64{123, 456}, []uint64{123, 456}, []uint64{123, 456}},
	}

	for _, test := range tests {
		s.systemd.Data.Statuses = make(map[string]systemd.UnitStatus)
		for _, bundle := range test.local {
			for i := 0; i < 3; i++ {
				serviceName := fmt.Sprintf("%d:%s", bundle, uuid.New())
				s.systemd.Data.Statuses[serviceName] = systemd.UnitStatus{UnitStatus: dbus.UnitStatus{Name: serviceName}}
			}
		}
		s.clusterConf.Data.Bundles = make(map[uint64]*clusterconf.Bundle)
		for _, id := range test.known {
			s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: &clusterconf.BundleConf{ID: id}}
		}
		bundles, err := s.statsPusher.getBundles()
		if !s.NoError(err, test.desc) {
			continue
		}
		bundleIDs := make(uint64s, 0, len(bundles))
		for _, bundle := range bundles {
			bundleIDs = append(bundleIDs, bundle.ID)
		}
		sort.Sort(uint64s(test.result))
		sort.Sort(uint64s(bundleIDs))
		s.Equal(test.result, bundleIDs, test.desc)
	}
}

func (s *StatsPusher) TestRunHealthChecks() {
	// TODO: Write proper tests when this is done
	bundles := []*clusterconf.Bundle{
		{BundleConf: &clusterconf.BundleConf{ID: 123}},
	}

	healthy, err := s.statsPusher.runHealthChecks(bundles)
	s.NoError(err)
	s.Len(healthy, len(bundles))
}

func (s *StatsPusher) TestSendBundleHeartbeats() {
	serial := "foobar"
	ip := net.ParseIP("123.123.123.123")
	bundles := []uint64{123, 456}
	for _, id := range bundles {
		s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: &clusterconf.BundleConf{ID: id}}
	}
	s.NoError(s.statsPusher.sendBundleHeartbeats(bundles, serial, ip))
	for _, id := range bundles {
		s.Equal(ip, s.clusterConf.Data.Bundles[id].Nodes[serial])
	}
}
