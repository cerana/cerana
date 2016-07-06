package main

import (
	"errors"
	"net"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/providers/zfs"
)

func (s *statsPusher) datasetHeartbeats() error {
	datasets, err := s.getDatasets()
	if err != nil {
		return err
	}
	ip, err := s.getIP()
	if err != nil {
		return err
	}

	return s.sendDatasetHeartbeats(datasets, ip)
}

func (s *statsPusher) getDatasets() ([]string, error) {
	ch := make(chan *acomm.Response, 1)
	defer close(ch)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		ch <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "zfs-list",
		ResponseHook: s.tracker.URL(),
		Args: zfs.ListArgs{
			Name: s.config.datasetDir(),
		},
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, err
	}

	if err := s.tracker.TrackRequest(req, s.config.requestTimeout()); err != nil {
		return nil, err
	}
	if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
		_ = s.tracker.RemoveRequest(req)
		return nil, err
	}

	resp := <-ch
	if resp.Error != nil {
		return nil, resp.Error
	}

	var listResult zfs.ListResult
	if err := resp.UnmarshalResult(&listResult); err != nil {
		return nil, err
	}

	// extract just the dataset ids and remove the base directory
	datasetIDs := make([]string, 0, len(listResult.Datasets))
	for _, dataset := range listResult.Datasets {
		if s.config.datasetDir() == dataset.Name {
			continue
		}
		datasetIDs = append(datasetIDs, filepath.Base(dataset.Name))
	}

	return datasetIDs, nil
}

func (s *statsPusher) getIP() (net.IP, error) {
	doneChan := make(chan *acomm.Response, 1)
	defer close(doneChan)

	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "metrics-network",
		ResponseHook:   s.tracker.URL(),
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, err
	}
	if err := s.tracker.TrackRequest(req, s.config.requestTimeout()); err != nil {
		return nil, err
	}
	if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
		return nil, err
	}

	resp := <-doneChan
	if resp.Error != nil {
		return nil, resp.Error
	}

	var data metrics.NetworkResult
	if err := resp.UnmarshalResult(&data); err != nil {
		return nil, err
	}
	for _, iface := range data.Interfaces {
		for _, ifaceAddr := range iface.Addrs {
			ip := net.ParseIP(ifaceAddr.Addr)
			if ip != nil && !ip.IsLoopback() {
				return ip, nil
			}
		}
	}
	return nil, errors.New("no suitable IP found")
}

func (s *statsPusher) sendDatasetHeartbeats(datasets []string, ip net.IP) error {
	var errored bool
	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for _, dataset := range datasets {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "dataset-heartbeat",
			Args: clusterconf.DatasetHeartbeatArgs{
				ID: dataset,
				IP: ip,
			},
		})
		if err != nil {
			errored = true
			continue
		}
		if err := multiRequest.AddRequest(dataset, req); err != nil {
			errored = true
			continue
		}
		if err := acomm.Send(s.config.clusterDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errored = true
			continue
		}
	}
	responses := multiRequest.Responses()
	for _, resp := range responses {
		if resp.Error != nil {
			errored = true
			break
		}
	}

	if errored {
		return errors.New("one or more dataset heartbeats unsuccessful")
	}
	return nil
}
