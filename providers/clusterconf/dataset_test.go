package clusterconf_test

import (
	"net"
	"path"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/pborman/uuid"
)

func (s *clusterConf) TestGetDataset() {
	dataset, err := s.addDataset()
	s.Require().NoError(err)

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
		}
	}
}

func (s *clusterConf) TestUpdateDataset() {
	dataset, err := s.addDataset()
	s.Require().NoError(err)
	dataset2, err := s.addDataset()
	s.Require().NoError(err)

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
					ID:       test.id,
					ModIndex: test.modIndex,
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
	dataset, err := s.addDataset()
	s.Require().NoError(err)

	tests := []struct {
		id  string
		err string
	}{
		{"", "missing arg: id"},
		{"does-not-exist", ""},
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
	dataset, err := s.addDataset()
	s.Require().NoError(err)

	tests := []struct {
		id    string
		ip    net.IP
		inUse bool
		err   string
	}{
		{"", net.IP{}, false, "missing arg: id"},
		{"", net.ParseIP("127.0.0.2"), false, "missing arg: id"},
		{dataset.ID, net.IP{}, false, "missing arg: ip"},
		{uuid.New(), net.ParseIP("127.0.0.3"), false, ""},
		{dataset.ID, net.ParseIP("127.0.0.4"), false, ""},
		{dataset.ID, net.ParseIP("127.0.0.4"), true, ""},
	}

	for _, test := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "dataset-heartbeat",
			Args: &clusterconf.DatasetHeartbeatArgs{ID: test.id, IP: test.ip},
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
			s.Nil(result, args)
		}
	}
}

func (s *clusterConf) TestListDatasetHeartbeats() {
	id := uuid.New()
	hb := clusterconf.DatasetHeartbeat{
		IP:    net.ParseIP("123.123.123.123"),
		InUse: true,
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "dataset-heartbeat",
		Args: &clusterconf.DatasetHeartbeatArgs{ID: id, IP: hb.IP, InUse: hb.InUse},
	})
	s.Require().NoError(err)
	_, _, err = s.clusterConf.DatasetHeartbeat(req)
	s.Require().NoError(err)

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "dataset-list-heartbeats",
	})
	s.Require().NoError(err)
	result, streamURL, err := s.clusterConf.ListDatasetHeartbeats(req)
	s.NoError(err)
	s.Nil(streamURL)
	if !s.NotNil(result) {
		return
	}
	hbList := result.(clusterconf.DatasetHeartbeatList)
	dsHBs, ok := hbList.Heartbeats[id]
	if !s.True(ok) {
		return
	}
	if !s.Len(dsHBs, 1) {
		return
	}
	s.EqualValues(hb, dsHBs[hb.IP.String()])
}

func (s *clusterConf) addDataset() (*clusterconf.Dataset, error) {
	dataset := &clusterconf.Dataset{
		ID:    uuid.New(),
		Quota: 5,
	}
	datasetKey := path.Join("datasets", dataset.ID, "config")

	data := map[string]interface{}{
		datasetKey: dataset,
	}
	indexes, err := s.loadData(data)
	if err != nil {
		return nil, err
	}

	dataset.ModIndex = indexes[datasetKey]

	return dataset, nil
}
