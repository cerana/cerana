package main

import (
	"fmt"
	"net"
	"sort"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/coreos/go-systemd/dbus"
	"github.com/pborman/uuid"
)

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
		known  []int
		local  []int
		result []int
	}{
		{"empty", []int{}, []int{}, []int{}},
		{"known only", []int{123}, []int{}, []int{}},
		{"local only", []int{}, []int{123}, []int{}},
		{"both, single", []int{123}, []int{123}, []int{123}},
		{"extra local", []int{123}, []int{123, 456}, []int{123}},
		{"extra known", []int{123, 456}, []int{123}, []int{123}},
		{"both, multiple", []int{123, 456}, []int{123, 456}, []int{123, 456}},
	}

	for _, test := range tests {
		s.systemd.Data.Statuses = make(map[string]dbus.UnitStatus)
		for _, bundle := range test.local {
			for i := 0; i < 3; i++ {
				serviceName := fmt.Sprintf("%d:%s", bundle, uuid.New())
				s.systemd.Data.Statuses[serviceName] = dbus.UnitStatus{Name: serviceName}
			}
		}
		s.clusterConf.Data.Bundles = make(map[int]*clusterconf.Bundle)
		for _, id := range test.known {
			s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: &clusterconf.BundleConf{ID: id}}
		}
		bundles, err := s.statsPusher.getBundles()
		if !s.NoError(err, test.desc) {
			continue
		}
		bundleIDs := make([]int, 0, len(bundles))
		for _, bundle := range bundles {
			bundleIDs = append(bundleIDs, bundle.ID)
		}
		sort.Ints(test.result)
		sort.Ints(bundleIDs)
		s.Equal(test.result, bundleIDs, test.desc)
	}
}

func (s *StatsPusher) TestRunHealthChecks() {
	// TODO: Write proper tests when this is done
	bundles := []*clusterconf.Bundle{
		&clusterconf.Bundle{BundleConf: &clusterconf.BundleConf{ID: 123}},
	}

	healthy, err := s.statsPusher.runHealthChecks(bundles)
	s.NoError(err)
	s.Len(healthy, len(bundles))
}

func (s *StatsPusher) TestSendBundleHeartbeats() {
	serial := "foobar"
	ip := net.ParseIP("123.123.123.123")
	bundles := []int{123, 456}
	for _, id := range bundles {
		s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: &clusterconf.BundleConf{ID: id}}
	}
	s.NoError(s.statsPusher.sendBundleHeartbeats(bundles, serial, ip))
	for _, id := range bundles {
		s.Equal(ip, s.clusterConf.Data.Bundles[id].Nodes[serial])
	}
}
