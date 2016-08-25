package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/cerana/cerana/tick"
)

func datasetTick(c tick.Configer, tracker *acomm.Tracker) error {
	config, ok := c.(*Config)
	if !ok {
		return errors.New("not the right type of config")
	}

	nodes, datasetHeartbeats, err := getCurrentState(config, tracker)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"nodes":             nodes,
		"datasetHeartbeats": datasetHeartbeats,
	}).Debug("current state")

	return replicateDatasets(config, tracker, nodes, datasetHeartbeats)
}

func getCurrentState(config *Config, tracker *acomm.Tracker) ([]clusterconf.Node, map[string]map[string]clusterconf.DatasetHeartbeat, error) {
	tasks := map[string]interface{}{
		"list-dataset-heartbeats": &clusterconf.DatasetHeartbeatList{},
		"list-datasets":           &clusterconf.DatasetListResult{},
		"list-nodes":              &clusterconf.ListNodesResult{},
	}
	// TODO: Use the config.HTTPResponseURL() for responsehook so that it can be run from anywhere
	multirequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for task := range tasks {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: task})
		if err != nil {
			return nil, nil, err
		}
		if err := multirequest.AddRequest(task, req); err != nil {
			return nil, nil, err
		}
		if err := acomm.Send(config.ClusterDataURL(), req); err != nil {
			multirequest.RemoveRequest(req)
			return nil, nil, err
		}
	}

	responses := multirequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			return nil, nil, errors.Wrapv(resp.Error, map[string]interface{}{"task": name})
		}
		if err := resp.UnmarshalResult(tasks[name]); err != nil {
			return nil, nil, errors.Wrapv(err, map[string]interface{}{"task": name})
		}
	}

	heartbeats := tasks["list-dataset-heartbeats"].(*clusterconf.DatasetHeartbeatList).Heartbeats
	datasets := tasks["list-datasets"].(*clusterconf.DatasetListResult).Datasets
	nodes := tasks["list-nodes"].(*clusterconf.ListNodesResult).Nodes

	// map for quick lookup of known datasets by ID
	datasetMap := make(map[string]struct{})
	for _, dataset := range datasets {
		datasetMap[dataset.ID] = struct{}{}
	}

	// remove any datasets not configured by clusterconf
	for datasetID := range heartbeats {
		if _, ok := datasetMap[datasetID]; !ok {
			delete(heartbeats, datasetID)
		}
	}

	return nodes, heartbeats, nil
}

func replicateDatasets(config *Config, tracker *acomm.Tracker, nodes []clusterconf.Node, datasets map[string]map[string]clusterconf.DatasetHeartbeat) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(datasets)*len(nodes))

	for datasetID, heartbeats := range datasets {
		replicateN := len(nodes) - len(heartbeats)
		if replicateN == 0 {
			// already present on all nodes
			logrus.WithField("dataset", datasetID).Debug("dataset already present on all nodes")
			continue
		}

		// determine a source node for the dataset and all required destination nodes
		var sourceIP string
		destinationIPs := make([]string, 0, replicateN)
		for _, node := range nodes {
			if _, ok := heartbeats[node.ID]; ok {
				if sourceIP == "" {
					sourceIP = node.ID
				}
			} else {
				destinationIPs = append(destinationIPs, node.ID)
			}
		}

		logrus.WithFields(logrus.Fields{
			"dataset":      datasetID,
			"source":       sourceIP,
			"destinations": destinationIPs,
		}).Debug("required replications")

		// send from the sourceIP and receive on each destination IP
		for _, destinationIP := range destinationIPs {
			replicateDataset(config, tracker, &wg, errorChan, datasetID, sourceIP, destinationIP)
		}
	}

	wg.Wait()

	// determine whether any replications failed
	var err error
	close(errorChan)
	if len(errorChan) > 0 {
		errs := make([]error, 0, len(errorChan))
		for e := range errorChan {
			errs = append(errs, e)
		}
		err = errors.Newv("not all datasets replicated", map[string]interface{}{"replicationErrors": errs})
	}

	return err
}

func replicateDataset(config *Config, tracker *acomm.Tracker, wg *sync.WaitGroup, errorChan chan error, datasetID, sourceIP, destinationIP string) {
	wg.Add(1)

	trackError := genTrackReplicationError(errorChan, wg, datasetID, sourceIP, destinationIP)
	datasetName := filepath.Join(config.DatasetPrefix(), datasetID)
	snapshotName := fmt.Sprintf("%s@%s", datasetName, datasetID)

	opts := acomm.RequestOptions{
		Task:         "zfs-send",
		ResponseHook: config.HTTPResponseURL(),
		Args:         zfs.CommonArgs{Name: snapshotName},
		ErrorHandler: func(req *acomm.Request, resp *acomm.Response) {
			trackError(req.Task, resp.Error)
		},
		SuccessHandler: func(req *acomm.Request, resp *acomm.Response) {
			opts := acomm.RequestOptions{
				Task:         "zfs-receive",
				ResponseHook: config.HTTPResponseURL(),
				Args:         zfs.CommonArgs{Name: datasetName},
				StreamURL:    resp.StreamURL,
				ErrorHandler: func(req *acomm.Request, resp *acomm.Response) {
					trackError(req.Task, resp.Error)
				},
				SuccessHandler: func(req *acomm.Request, resp *acomm.Response) {
					logrus.WithFields(logrus.Fields{
						"dataset":     datasetID,
						"source":      sourceIP,
						"destination": destinationIP,
					}).Debug("dataset replication successful")
					wg.Done()
				},
			}

			if err := tick.SendNodeRequest(config, tracker, opts, destinationIP); err != nil {
				trackError(opts.Task, err)
			}
		},
	}

	if err := tick.SendNodeRequest(config, tracker, opts, sourceIP); err != nil {
		trackError(opts.Task, err)
	}
}

func genTrackReplicationError(errorChan chan error, wg *sync.WaitGroup, id, source, destination string) func(string, error) {
	return func(task string, err error) {
		errorChan <- errors.Wrapv(err, map[string]interface{}{
			"task":        task,
			"dataset":     id,
			"source":      source,
			"destination": destination,
		})

		wg.Done()
	}
}
