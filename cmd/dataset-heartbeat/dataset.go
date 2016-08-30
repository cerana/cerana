package main

import (
	"net"
	"net/url"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/cerana/cerana/tick"
)

func datasetHeartbeats(config tick.Configer, tracker *acomm.Tracker) error {
	conf, ok := config.(*Config)
	if !ok {
		return errors.New("not the right type of config")
	}

	ip, err := tick.GetIP(conf, tracker)
	if err != nil {
		return err
	}

	datasets, err := getDatasets(conf, tracker, ip)
	if err != nil {
		return err
	}

	return sendDatasetHeartbeats(conf, tracker, datasets, ip)
}

func getDatasets(config *Config, tracker *acomm.Tracker, ip net.IP) ([]clusterconf.DatasetHeartbeatArgs, error) {
	requests := map[string]struct {
		task        string
		coordinator *url.URL
		args        interface{}
		respData    interface{}
	}{
		"datasets":  {task: "zfs-list", coordinator: config.NodeDataURL(), args: zfs.ListArgs{Name: config.DatasetPrefix()}, respData: &zfs.ListResult{}},
		"bundles":   {task: "list-bundles", coordinator: config.ClusterDataURL(), respData: &clusterconf.BundleListResult{}},
		"bundleHBs": {task: "list-bundle-heartbeats", coordinator: config.ClusterDataURL(), respData: &clusterconf.BundleHeartbeatList{}},
	}

	multiRequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for name, args := range requests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: args.task,
			Args: args.args,
		})
		if err != nil {
			return nil, err
		}
		if err := multiRequest.AddRequest(name, req); err != nil {
			return nil, err
		}
		if err := acomm.Send(args.coordinator, req); err != nil {
			multiRequest.RemoveRequest(req)
			return nil, err
		}
	}

	responses := multiRequest.Responses()
	for name, args := range requests {
		resp := responses[name]
		if resp.Error != nil {
			return nil, errors.ResetStack(resp.Error)
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
		if config.DatasetPrefix() == dataset.Name {
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

func sendDatasetHeartbeats(config *Config, tracker *acomm.Tracker, datasetArgs []clusterconf.DatasetHeartbeatArgs, ip net.IP) error {
	var errored bool
	multiRequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
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
		if err := acomm.Send(config.ClusterDataURL(), req); err != nil {
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
