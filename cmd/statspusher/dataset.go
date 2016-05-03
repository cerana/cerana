package main

import (
	"errors"
	"fmt"
	"net"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/mistifyio/go-zfs"
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
	datasets := []string{}

	requests := make(map[string]*acomm.Request)
	localReq, err := acomm.NewRequest(acomm.RequestOptions{Task: "zfs-list"})
	if err != nil {
		return datasets, err
	}
	requests["local"] = localReq
	knownReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task:         "get-datasets",
		HeartbeatURL: s.config.heartbeatURL(),
	})
	if err != nil {
		return datasets, err
	}
	requests["known"] = knownReq

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			break
		}
		if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			break
		}
	}

	responses := multiRequest.Responses()
	for name := range requests {
		resp, ok := responses[name]
		if !ok {
			return datasets, fmt.Errorf("failed to send request: %s", name)
		}
		if resp.Error != nil {
			return datasets, fmt.Errorf("request failed: %s: %s", name, resp.Error)
		}
		if err := resp.UnmarshalResult(tasks[name]); err != nil {
			return datasets, err
		}
	}

	localDatasets := responses["local"].([]*zfs.Dataset)
	knownDatasets := responses["known"].([]*zfs.Dataset)
	datasets = make([]string, 0, len(localDatasets))
	for _, known := range localDatasets {
		for _, local := range knownDatasets {
			if known.ID == local.Name {
				datasets = append(datasets, local.Name)
				break
			}
		}
	}
	return datasets, nil
}

func (s *statsPusher) getIP() (*net.IP, error) {
	doneChan := make(chan *acomm.Response, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "metrics-network",
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Send(s.config.coordinatorURL()); err != nil {
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

func (s *statsPusher) sendDatasetHeartbeats(datasets []*zfs.Dataset, ip *net.IP) error {
	var errored bool

	multiRequest := acomm.NewMultirequest(s.tracker, s.config.requestTimeout())
	for _, dataset := range datasets {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task:    "dataset-heartbeat",
			TaskURL: s.config.heartbeatURL(),
			Args: DatasetHeartbeatArgs{
				ID: dataset.Name,
				IP: ip,
			},
		})
		if err != nil {
			errored = true
			continue
		}
		if err := multirequest.AddRequest(dataset.Name, req); err != nil {
			errored = true
			continue
		}
		if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			errored = true
			continue
		}
	}
	responses := multiRequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			errored = true
		}
	}

	if errored {
		return errors.New("one or more dataset heartbeats unsuccessful")
	}
	return nil
}
