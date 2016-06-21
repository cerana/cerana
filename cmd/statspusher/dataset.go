package main

import (
	"errors"
	"net"
	"net/url"

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
	requests := map[string]struct {
		task     string
		url      *url.URL
		respData interface{}
	}{
		"local": {task: "zfs-list", url: s.config.nodeDataURL(), respData: &zfs.ListResult{}},
		"known": {task: "list-datasets", url: s.config.clusterDataURL(), respData: &clusterconf.DatasetListResult{}},
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
		if err := acomm.Send(args.url, req); err != nil {
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
	localDatasets := requests["local"].respData.(*zfs.ListResult).Datasets
	knownDatasets := requests["known"].respData.(*clusterconf.DatasetListResult).Datasets

	datasets := make([]string, 0, len(localDatasets))
	for _, local := range localDatasets {
		for _, known := range knownDatasets {
			if known.ID == local.Name {
				datasets = append(datasets, local.Name)
				break
			}
		}
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
