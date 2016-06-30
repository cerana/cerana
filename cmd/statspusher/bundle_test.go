package main

import (
	"fmt"
	"net"
	"sort"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/health"
	"github.com/cerana/cerana/providers/systemd"
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
		s.systemd.ClearData()
		for _, bundle := range test.local {
			for i := 0; i < 3; i++ {
				serviceName := fmt.Sprintf("%d:%s.service", bundle, uuid.New())
				s.systemd.ManualCreate(systemd.CreateArgs{
					Name: serviceName,
				}, true)
			}
		}
		s.clusterConf.Data.Bundles = make(map[uint64]*clusterconf.Bundle)
		for _, id := range test.known {
			s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: clusterconf.BundleConf{ID: id}}
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
	s.health.Data.Uptime = false
	bundle := &clusterconf.Bundle{
		BundleConf: clusterconf.BundleConf{
			ID: 123,
			Services: map[string]clusterconf.BundleService{
				"foobar": {
					ServiceConf: clusterconf.ServiceConf{
						HealthChecks: map[string]clusterconf.HealthCheck{
							"file": {
								ID:   "file",
								Type: "health-file",
								Args: health.FileArgs{},
							},
							"uptime": {
								ID:   "file",
								Type: "health-uptime",
								Args: health.UptimeArgs{},
							},
						},
					},
				},
			},
		},
	}
	bundles := []*clusterconf.Bundle{bundle}

	healthErrors, err := s.statsPusher.runHealthChecks(bundles)
	s.Len(err, 0)
	b, ok := healthErrors[bundle.ID]
	if !s.True(ok) {
		return
	}
	_, ok = b["foobar:file"]
	s.False(ok)
	_, ok = b["foobar:uptime"]
	s.True(ok)
}

func (s *StatsPusher) TestSendBundleHeartbeats() {
	serial := "foobar"
	ip := net.ParseIP("123.123.123.123")
	bundles := map[uint64]map[string]error{
		123: {},
		456: {},
	}
	for id := range bundles {
		s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{BundleConf: clusterconf.BundleConf{ID: id}}
	}
	s.NoError(s.statsPusher.sendBundleHeartbeats(bundles, serial, ip))
	for id := range bundles {
		s.Equal(ip, s.clusterConf.Data.Bundles[id].Nodes[serial].IP)
	}
}
