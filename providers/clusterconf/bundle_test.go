package clusterconf_test

import (
	"encoding/json"
	"errors"
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
		desc            string
		id              uint64
		quota           uint64
		dataset         string
		combinedOverlay bool
		err             string
	}{
		{"zero id", 0, 0, "", false, "missing arg: id"},
		{"nonexistent id", uint64(rand.Int63()), 0, "", false, "bundle config not found"},
		{"existent id config", bundle.ID, 0, "", false, ""},
		{"existent id overlayed", bundle.ID, 5, "testds", true, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "get-bundle",
			Args: &clusterconf.GetBundleArgs{
				ID:              test.id,
				CombinedOverlay: test.combinedOverlay,
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.GetBundle(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.Contains(err.Error(), test.err, test.desc)
			s.Nil(result, test.desc)
		} else {
			s.NoError(err, test.desc)
			if !s.NotNil(result, test.desc) {
				continue
			}
			bundlePayload, ok := result.(*clusterconf.BundlePayload)
			s.True(ok, test.desc)
			s.Equal(test.id, bundlePayload.Bundle.ID, test.desc)
			for _, ds := range bundlePayload.Bundle.Datasets {
				s.Equal(test.quota, ds.Quota, test.desc)
			}
			for _, service := range bundlePayload.Bundle.Services {
				s.Equal(test.dataset, service.Dataset, test.desc)
			}
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
					ID:       test.id,
					ModIndex: test.modIndex,
				},
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.UpdateBundle(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.Contains(err.Error(), test.err, test.desc)
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
			Args: &clusterconf.DeleteBundleArgs{ID: test.id},
		})
		s.Require().NoError(err, desc)
		result, streamURL, err := s.clusterConf.DeleteBundle(req)
		s.Nil(streamURL, desc)
		s.Nil(result, desc)
		if test.err != "" {
			s.Contains(err.Error(), test.err, desc)
		} else {
			s.NoError(err, desc)
		}
	}
}

func (s *clusterConf) TestBundleHeartbeat() {
	tests := []struct {
		id     uint64
		serial string
		ip     net.IP
		err    string
	}{
		{0, "", net.IP{}, "missing arg: id"},
		{0, "", net.ParseIP("127.0.0.2"), "missing arg: id"},
		{uint64(rand.Int63()), "", net.IP{}, "missing arg: serial"},
		{uint64(rand.Int63()), uuid.New(), net.IP{}, "missing arg: ip"},
		{uint64(rand.Int63()), uuid.New(), net.ParseIP("127.0.0.4"), ""},
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
			s.Contains(err.Error(), test.err, args)
			s.Nil(result, args)
		} else {
			s.NoError(err, args)
			s.Nil(result, args)
		}
	}
}

func (s *clusterConf) TestBundleHeartbeatJSON() {
	heartbeatList := clusterconf.BundleHeartbeatList{
		Heartbeats: map[uint64]clusterconf.BundleHeartbeats{
			uint64(1): {
				uuid.New(): clusterconf.BundleHeartbeat{
					IP: net.ParseIP("192.168.1.1"),
					HealthErrors: map[string]error{
						uuid.New(): errors.New("test"),
					},
				},
			},
		},
	}
	j, err := json.Marshal(heartbeatList)
	if !s.NoError(err) {
		return
	}

	var heartbeatList2 clusterconf.BundleHeartbeatList
	s.NoError(json.Unmarshal(j, &heartbeatList2))

	s.Equal(heartbeatList, heartbeatList2)
}

func (s *clusterConf) TestListBundleHeartbeats() {
	id := uint64(5)
	serial := uuid.New()
	hb := clusterconf.BundleHeartbeat{
		IP:           net.ParseIP("123.123.123.123"),
		HealthErrors: make(map[string]error),
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "dataset-heartbeat",
		Args: &clusterconf.BundleHeartbeatArgs{ID: id, Serial: serial, IP: hb.IP},
	})
	s.Require().NoError(err)
	_, _, err = s.clusterConf.BundleHeartbeat(req)
	s.Require().NoError(err)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "bundle-list-heartbeats",
	})
	s.Require().NoError(err)
	result, streamURL, err := s.clusterConf.ListBundleHeartbeats(req)
	s.NoError(err)
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
	hbList := result.(clusterconf.BundleHeartbeatList)
	dsHBs, ok := hbList.Heartbeats[id]
	if !s.True(ok) {
		return
	}
	if !s.Len(dsHBs, 1) {
		return
	}
	s.EqualValues(hb, dsHBs[serial])
}

func (s *clusterConf) addBundle() (*clusterconf.Bundle, error) {
	service, err := s.addService()
	if err != nil {
		return nil, err
	}
	dataset, err := s.addDataset()
	if err != nil {
		return nil, err
	}
	bundle := &clusterconf.Bundle{
		ID: uint64(rand.Int63()),
		Datasets: map[string]clusterconf.BundleDataset{
			dataset.ID: {
				Name: "foobar",
				ID:   dataset.ID,
			},
		},
		Services: map[string]clusterconf.BundleService{
			service.ID: {
				ServiceConf: clusterconf.ServiceConf{ID: service.ID},
			},
		},
		Ports: clusterconf.BundlePorts{
			1: clusterconf.BundlePort{
				Port: 1,
			},
		},
	}
	bundleKey := path.Join("bundles", strconv.FormatUint(bundle.ID, 10), "config")

	data := map[string]interface{}{
		bundleKey: bundle,
	}
	indexes, err := s.loadData(data)
	if err != nil {
		return nil, err
	}

	bundle.ModIndex = indexes[bundleKey]

	return bundle, nil
}
