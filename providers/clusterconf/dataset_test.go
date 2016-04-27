package clusterconf_test

import (
	"encoding/json"
	"net"
	"path"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/mistifyio/lochness/pkg/kv"
	"github.com/pborman/uuid"
)

func (s *clusterConf) TestGetDataset() {
	dataset := s.addDataset()

	tests := []struct {
		id    string
		nodes map[string]bool
		err   string
	}{
		{"", make(map[string]bool), "missing arg: id"},
		{"does-not-exist", make(map[string]bool), "dataset config not found"},
		{dataset.ID, dataset.Nodes, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "get-dataset",
			Args: &clusterconf.IDArgs{ID: test.id},
		})
		s.Require().NoError(err, test.id)
		result, streamURL, err := s.clusterConf.GetDataset(req)
		s.Nil(streamURL, test.id)
		if test.err != "" {
			s.EqualError(err, test.err, test.id)
			s.Nil(result, test.id)
		} else {
			s.NoError(err, test.id)
			if !s.NotNil(result, test.id) {
				continue
			}
			datasetPayload, ok := result.(*clusterconf.DatasetPayload)
			s.True(ok, test.id)
			s.Equal(test.id, datasetPayload.Dataset.ID, test.id)
			s.Equal(test.nodes, datasetPayload.Dataset.Nodes, test.id)
		}
	}
}

func (s *clusterConf) TestUpdateDataset() {
	dataset := s.addDataset()
	dataset2 := s.addDataset()
	tests := []struct {
		desc     string
		id       string
		modIndex uint64
		err      string
	}{
		{"no id", "", 0, ""},
		{"new id", uuid.New(), 0, ""},
		{"create existing id", dataset.ID, 0, "CAS failed"},
		{"update existing id", dataset2.ID, dataset2.ModIndex, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "update-dataset",
			Args: &clusterconf.DatasetPayload{
				Dataset: &clusterconf.Dataset{
					DatasetConf: &clusterconf.DatasetConf{ID: test.id},
					ModIndex:    test.modIndex,
				},
			},
		})
		s.Require().NoError(err, test.desc)
		result, streamURL, err := s.clusterConf.UpdateDataset(req)
		s.Nil(streamURL, test.desc)
		if test.err != "" {
			s.EqualError(err, test.err, test.desc)
			s.Nil(result, test.desc)
		} else {
			s.NoError(err, test.desc)
			if !s.NotNil(result, test.desc) {
				continue
			}
			datasetPayload, ok := result.(*clusterconf.DatasetPayload)
			s.True(ok, test.desc)
			if test.id == "" {
				s.NotEmpty(datasetPayload.Dataset.ID, test.desc)
			} else {
				s.Equal(test.id, datasetPayload.Dataset.ID, test.desc)
			}
			s.NotEqual(test.modIndex, datasetPayload.Dataset.ModIndex, test.desc)
		}
	}
}

func (s *clusterConf) TestDeleteDataset() {
	dataset := s.addDataset()

	tests := []struct {
		id  string
		err string
	}{
		{"", "missing arg: id"},
		{"does-not-exist", "dataset config not found"},
		{dataset.ID, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "delete-dataset",
			Args: &clusterconf.IDArgs{ID: test.id},
		})
		s.Require().NoError(err, test.id)
		result, streamURL, err := s.clusterConf.DeleteDataset(req)
		s.Nil(streamURL, test.id)
		s.Nil(result, test.id)
		if test.err != "" {
			s.EqualError(err, test.err, test.id)
		} else {
			s.NoError(err, test.id)
		}
	}
}

func (s *clusterConf) TestDatasetHeartbeat() {
	dataset := s.addDataset()

	tests := []struct {
		id  string
		ip  net.IP
		err string
	}{
		{"", net.IP{}, "missing arg: id"},
		{"", net.ParseIP("127.0.0.2"), "missing arg: id"},
		{dataset.ID, net.IP{}, "missing arg: ip"},
		{uuid.New(), net.ParseIP("127.0.0.3"), "dataset config not found"},
		{dataset.ID, net.ParseIP("127.0.0.4"), ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "dataset-heartbeat",
			Args: &clusterconf.DatasetHeartbeatArgs{test.id, test.ip},
		})
		args := string(*req.Args)
		s.Require().NoError(err, args)
		result, streamURL, err := s.clusterConf.DatasetHeartbeat(req)
		s.Nil(streamURL, args)
		if test.err != "" {
			s.EqualError(err, test.err, args)
			s.Nil(result, args)
		} else {
			s.NoError(err, args)
			if !s.NotNil(result, args) {
				continue
			}
			datasetPayload, ok := result.(*clusterconf.DatasetPayload)
			s.True(ok, args)
			if test.id == "" {
				s.NotEmpty(datasetPayload.Dataset.ID, args)
			} else {
				s.Equal(test.id, datasetPayload.Dataset.ID, args)
			}
			s.True(datasetPayload.Dataset.Nodes[test.ip.String()], args)
		}
	}
}

func (s *clusterConf) addDataset() *clusterconf.Dataset {
	// Populate a dataset
	dataset := &clusterconf.Dataset{DatasetConf: &clusterconf.DatasetConf{ID: uuid.New()}}
	sj, _ := json.Marshal(dataset)
	key := path.Join("datasets", dataset.ID, "config")
	s.kvp.Data[key] = kv.Value{Data: sj, Index: 1}
	dataset.ModIndex = 1

	// Give it a node heartbeat
	key = path.Join("datasets", dataset.ID, "nodes", "127.0.0.1")
	hbval, _ := json.Marshal(true)
	s.kvp.Data[key] = kv.Value{Data: hbval, Index: 1}
	dataset.Nodes = map[string]bool{"127.0.0.1": true}
	return dataset
}
