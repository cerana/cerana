package main

import (
	"net"
	"sort"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/health"
	"github.com/cerana/cerana/providers/service"
	"github.com/pborman/uuid"
)

type uint64s []uint64

func (u uint64s) Len() int           { return len(u) }
func (u uint64s) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u uint64s) Less(i, j int) bool { return u[i] < u[j] }

func (s *BundleHeartbeat) TestGetSerial() {
	serial, err := getSerial(s.config, s.tracker)
	if !s.NoError(err) {
		return
	}
	s.Equal(s.metrics.Data.Host.Hostname, serial)
}

func (s *BundleHeartbeat) extractBundles() {

}

func (s *BundleHeartbeat) TestGetBundles() {
	tests := []struct {
		desc   string
		known  []uint64
		local  []uint64
		result uint64s
	}{
		{"empty", []uint64{}, []uint64{}, []uint64{}},
		{"known only", []uint64{123}, []uint64{}, []uint64{}},
		{"local only", []uint64{}, []uint64{123}, []uint64{123}},
		{"both, single", []uint64{123}, []uint64{123}, []uint64{123}},
		{"extra local", []uint64{123}, []uint64{123, 456}, []uint64{123, 456}},
		{"extra known", []uint64{123, 456}, []uint64{123}, []uint64{123}},
		{"both, multiple", []uint64{123, 456}, []uint64{123, 456}, []uint64{123, 456}},
	}

	for _, test := range tests {
		s.service.ClearData()
		for _, bundle := range test.local {
			for i := 0; i < 3; i++ {
				s.service.Add(service.Service{
					ID:       uuid.New(),
					BundleID: bundle,
				})
			}
		}
		s.clusterConf.Data.Bundles = make(map[uint64]*clusterconf.Bundle)
		for _, id := range test.known {
			s.clusterConf.Data.Bundles[id] = &clusterconf.Bundle{ID: id}
		}
		bundles, err := getBundles(s.config, s.tracker)
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

func (s *BundleHeartbeat) TestRunHealthChecks() {
	s.health.Data.Uptime = false
	bundle := &clusterconf.Bundle{
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
	}
	bundles := []*clusterconf.Bundle{bundle}

	healthErrors, err := runHealthChecks(s.config, s.tracker, bundles)
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

func (s *BundleHeartbeat) TestSendBundleHeartbeats() {
	serial := "foobar"
	ip := net.ParseIP("123.123.123.123")
	bundles := map[uint64]map[string]error{
		123: {},
		456: {},
	}
	s.NoError(sendBundleHeartbeats(s.config, s.tracker, bundles, serial, ip))
}
