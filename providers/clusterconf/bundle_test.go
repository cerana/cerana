package clusterconf_test

import (
	"math/rand"
	"net"
	"path"
	"strconv"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/pborman/uuid"
)

func (s *clusterConf) TestGetBundle() {
	bundle, err := s.addBundle()
	s.Require().NoError(err)

	tests := []struct {
		desc  string
		id    uint64
		nodes map[string]net.IP
		err   string
	}{
		{"zero id", 0, make(map[string]net.IP), "missing arg: id"},
		{"nonexistent id", uint64(rand.Int63()), make(map[string]net.IP), "bundle config not found"},
		{"existent id", bundle.ID, bundle.Nodes, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "get-bundle",
			Args: &clusterconf.BundleIDArgs{ID: test.id},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.GetBundle(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.EqualError(err, test.err, test.desc)
			s.Nil(result, test.desc)
		} else {
			s.NoError(err, test.desc)
			if !s.NotNil(result, test.desc) {
				continue
			}
			bundlePayload, ok := result.(*clusterconf.BundlePayload)
			s.True(ok, test.desc)
			s.Equal(test.id, bundlePayload.Bundle.ID, test.desc)
			s.Equal(test.nodes, bundlePayload.Bundle.Nodes, test.desc)
		}
	}
}

func (s *clusterConf) TestUpdateBundle() {
	bundle, err := s.addBundle()
	s.Require().NoError(err)
	bundle2, err := s.addBundle()
	s.Require().NoError(err)

	tests := []struct {
		desc     string
		id       uint64
		modIndex uint64
		err      string
	}{
		{"no id", 0, 0, ""},
		{"new id", uint64(rand.Int63()), 0, ""},
		{"create existing id", bundle.ID, 0, "CAS failed"},
		{"update existing id", bundle2.ID, bundle2.ModIndex, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "update-bundle",
			Args: &clusterconf.BundlePayload{
				Bundle: &clusterconf.Bundle{
					BundleConf: clusterconf.BundleConf{ID: test.id},
					ModIndex:   test.modIndex,
				},
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.UpdateBundle(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.EqualError(err, test.err, test.desc)
			s.Nil(result, test.desc)
		} else {
			s.NoError(err, test.desc)
			if !s.NotNil(result, test.desc) {
				continue
			}
			bundlePayload, ok := result.(*clusterconf.BundlePayload)
			s.True(ok, test.desc)
			if test.id == 0 {
				s.NotEmpty(bundlePayload.Bundle.ID, test.desc)
			} else {
				s.Equal(test.id, bundlePayload.Bundle.ID, test.desc)
			}
			s.NotEqual(test.modIndex, bundlePayload.Bundle.ModIndex, test.desc)
		}
	}
}

func (s *clusterConf) TestDeleteBundle() {
	bundle, err := s.addBundle()
	s.Require().NoError(err)

	tests := []struct {
		id  uint64
		err string
	}{
		{0, "missing arg: id"},
		{uint64(rand.Int63()), ""},
		{bundle.ID, ""},
	}

	for _, test := range tests {
		desc := strconv.FormatUint(test.id, 10)
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "delete-bundle",
			Args: &clusterconf.BundleIDArgs{ID: test.id},
		})
		s.Require().NoError(err, desc)
		result, streamURL, err := s.clusterConf.DeleteBundle(req)
		s.Nil(streamURL, desc)
		s.Nil(result, desc)
		if test.err != "" {
			s.EqualError(err, test.err, desc)
		} else {
			s.NoError(err, desc)
		}
	}
}

func (s *clusterConf) TestBundleHeartbeat() {
	bundle, err := s.addBundle()
	s.Require().NoError(err)

	tests := []struct {
		id     uint64
		serial string
		ip     net.IP
		err    string
	}{
		{0, "", net.IP{}, "missing arg: id"},
		{0, "", net.ParseIP("127.0.0.2"), "missing arg: id"},
		{bundle.ID, "", net.IP{}, "missing arg: serial"},
		{bundle.ID, uuid.New(), net.IP{}, "missing arg: ip"},
		{uint64(rand.Int63()), uuid.New(), net.ParseIP("127.0.0.3"), "bundle config not found"},
		{bundle.ID, uuid.New(), net.ParseIP("127.0.0.4"), ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "bundle-heartbeat",
			Args: &clusterconf.BundleHeartbeatArgs{
				ID:     test.id,
				Serial: test.serial,
				IP:     test.ip,
			},
		})
		args := string(*req.Args)
		s.Require().NoError(err, args)
		result, streamURL, err := s.clusterConf.BundleHeartbeat(req)
		s.Nil(streamURL, args)
		if test.err != "" {
			s.EqualError(err, test.err, args)
			s.Nil(result, args)
		} else {
			s.NoError(err, args)
			if !s.NotNil(result, args) {
				continue
			}
			bundlePayload, ok := result.(*clusterconf.BundlePayload)
			s.True(ok, args)
			if test.id == 0 {
				s.NotEmpty(bundlePayload.Bundle.ID, args)
			} else {
				s.Equal(test.id, bundlePayload.Bundle.ID, args)
			}
			s.NotEmpty(bundlePayload.Bundle.Nodes[test.serial], args)
		}
	}
}

func (s *clusterConf) addBundle() (*clusterconf.Bundle, error) {
	bundle := &clusterconf.Bundle{BundleConf: clusterconf.BundleConf{
		ID: uint64(rand.Int63()),
		Ports: clusterconf.BundlePorts{
			1: &clusterconf.BundlePort{
				Port: 1,
			},
		},
	}}
	bundleKey := path.Join("bundles", strconv.FormatUint(bundle.ID, 10), "config")

	serial := uuid.New()
	ip := net.ParseIP("127.0.0.1")
	nodeKey := path.Join("bundles", strconv.FormatUint(bundle.ID, 10), "nodes", serial)

	data := map[string]interface{}{
		bundleKey: bundle,
		nodeKey:   ip,
	}
	indexes, err := s.loadData(data)
	if err != nil {
		return nil, err
	}

	bundle.ModIndex = indexes[bundleKey]
	bundle.Nodes = map[string]net.IP{serial: ip}

	return bundle, nil
}
