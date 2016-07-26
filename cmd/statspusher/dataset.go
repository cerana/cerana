package main

import (
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/providers/zfs"
)

func (s *statsPusher) datasetHeartbeats() error {
	ip, err := s.getIP()
	if err != nil {
		return err
	}
	datasets, err := s.getDatasets(ip)
	if err != nil {
		return err
	}

	return s.sendDatasetHeartbeats(datasets, ip)
}

func (s *statsPusher) getDatasets(ip net.IP) ([]clusterconf.DatasetHeartbeatArgs, error) {
	requests := map[string]struct {
		task     string
		args     interface{}
		respData interface{}
	}{
		"datasets":  {task: "zfs-list", args: zfs.ListArgs{Name: s.config.datasetDir()}, respData: &zfs.ListResult{}},
		"bundles":   {task: "list-bundles", respData: &clusterconf.BundleListResult{}},
		"bundleHBs": {task: "list-bundle-heartbeats", respData: &clusterconf.BundleHeartbeatList{}},
	}

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for name, args := range requests {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: args.task})
		if err != nil {
			return nil, err
		}
		if err := multiRequest.AddRequest(name, req); err != nil {
			return nil, err
		}
		if err := acomm.Send(s.config.nodeDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			return nil, err
		}
	}

	responses := multiRequest.Responses()
	for name, args := range requests {
		resp := responses[name]
		if resp.Error != nil {
			return nil, resp.Error
		}
		if err := resp.UnmarshalResult(args.respData); err != nil {
			return nil, err
		}
	}

	listResult := requests["datasets"].respData.(*zfs.ListResult).Datasets
	bundles := requests["bundles"].respData.(*clusterconf.BundleListResult).Bundles
	heartbeats := requests["bundleHBs"].respData.(*clusterconf.BundleHeartbeatList).Heartbeats

	// determine which datasets are configured to be in use on this node
	datasetsInUse := make(map[string]bool)
	for _, bundle := range bundles {
		for _, hb := range heartbeats[bundle.ID] {
			if hb.IP.Equal(ip) {
				for datasetID := range bundle.Datasets {
					datasetsInUse[datasetID] = true
				}
				break
			}
		}
	}

	// extract just the dataset ids and ignore the base directory
	datasets := make([]clusterconf.DatasetHeartbeatArgs, 0, len(listResult))
	for _, dataset := range listResult {
		if s.config.datasetDir() == dataset.Name {
			continue
		}

		datasetID := filepath.Base(dataset.Name)

		args := clusterconf.DatasetHeartbeatArgs{
			ID:    datasetID,
			InUse: datasetsInUse[datasetID],
		}
		datasets = append(datasets, args)
	}

	return datasets, nil
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
			fmt.Println("Addr: " + ifaceAddr.Addr)
			ip, _, _ := net.ParseCIDR(ifaceAddr.Addr)
			if ip != nil && !ip.IsLoopback() {
				return ip, nil
			}
		}
	}
	return nil, errors.New("no suitable IP found")
}

func (s *statsPusher) sendDatasetHeartbeats(datasetArgs []clusterconf.DatasetHeartbeatArgs, ip net.IP) error {
	var errored bool
	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for _, dataset := range datasetArgs {
		dataset.IP = ip
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "dataset-heartbeat",
			Args: dataset,
		})
		if err != nil {
			errored = true
			continue
		}
		if err := multiRequest.AddRequest(dataset.ID, req); err != nil {
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
