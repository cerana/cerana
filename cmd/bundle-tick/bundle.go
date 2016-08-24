package main

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/tick"
)

func bundleTick(c tick.Configer, tracker *acomm.Tracker) error {
	config, ok := c.(*Config)
	if !ok {
		return errors.New("not the right type of config")
	}

	nodes, bundles, bundleHeartbeats, err := getCurrentState(config, tracker)
	if err != nil {
		return err
	}

	// annoyance since json marshal won't do map[uint64]
	heartbeatsForLogging := make(map[string]clusterconf.BundleHeartbeats)
	for id, hbs := range bundleHeartbeats {
		heartbeatsForLogging[fmt.Sprintf("%d", id)] = hbs
	}

	logrus.WithFields(logrus.Fields{
		"nodes":            nodes,
		"bundles":          bundles,
		"bundleHeartbeats": heartbeatsForLogging,
	}).Debug("current state")

	return runBundles(config, tracker, nodes, bundles, bundleHeartbeats)
}

func getCurrentState(config *Config, tracker *acomm.Tracker) ([]clusterconf.Node, []*clusterconf.Bundle, map[uint64]clusterconf.BundleHeartbeats, error) {
	tasks := map[string]struct {
		args   interface{}
		result interface{}
	}{
		"list-bundle-heartbeats": {nil, &clusterconf.BundleHeartbeatList{}},
		"list-bundles":           {clusterconf.ListBundleArgs{CombinedOverlay: true}, &clusterconf.BundleListResult{}},
		"list-nodes":             {nil, &clusterconf.ListNodesResult{}},
	}
	// TODO: Use the config.HTTPResponseURL() for responsehook so that it can be run from anywhere
	multirequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for task, taskMeta := range tasks {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: task, Args: taskMeta.args})
		if err != nil {
			return nil, nil, nil, err
		}
		if err := multirequest.AddRequest(task, req); err != nil {
			return nil, nil, nil, err
		}
		if err := acomm.Send(config.ClusterDataURL(), req); err != nil {
			multirequest.RemoveRequest(req)
			return nil, nil, nil, err
		}
	}

	responses := multirequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			return nil, nil, nil, errors.Wrapv(resp.Error, map[string]interface{}{"task": name})
		}
		if err := resp.UnmarshalResult(tasks[name].result); err != nil {
			return nil, nil, nil, errors.Wrapv(err, map[string]interface{}{"task": name})
		}
	}

	heartbeats := tasks["list-bundle-heartbeats"].result.(*clusterconf.BundleHeartbeatList).Heartbeats
	bundles := tasks["list-bundles"].result.(*clusterconf.BundleListResult).Bundles
	nodes := tasks["list-nodes"].result.(*clusterconf.ListNodesResult).Nodes

	// map for quick lookup of known bundles by ID
	bundleMap := make(map[uint64]struct{})
	for _, bundle := range bundles {
		bundleMap[bundle.ID] = struct{}{}
	}

	// remove any bundles not configured by clusterconf
	for bundleID := range heartbeats {
		if _, ok := bundleMap[bundleID]; !ok {
			delete(heartbeats, bundleID)
		}
	}

	return nodes, bundles, heartbeats, nil
}

func runBundles(config *Config, tracker *acomm.Tracker, nodes []clusterconf.Node, bundles []*clusterconf.Bundle, bundleHeartbeats map[uint64]clusterconf.BundleHeartbeats) error {
	var wg sync.WaitGroup
	// make error channel large enough to handle every service of every bundle on all nodes
	count := 0
	for _, b := range bundles {
		count += len(b.Services)
	}
	errorChan := make(chan error, count*len(nodes))

	for _, bundle := range bundles {
		heartbeats := bundleHeartbeats[bundle.ID]

		runN := len(nodes) - len(heartbeats)
		if runN == 0 {
			// already present on all nodes
			logrus.WithField("bundle", bundle.ID).Debug("bundle already present on all nodes")
			continue
		}

		nodeIPs := make([]string, 0, runN)
		for _, node := range nodes {
			if heartbeats == nil {
				nodeIPs = append(nodeIPs, node.ID)
			} else if _, ok := heartbeats[node.ID]; !ok {
				nodeIPs = append(nodeIPs, node.ID)
			}
		}

		logrus.WithFields(logrus.Fields{
			"bundle": bundle.ID,
			"nodes":  nodeIPs,
		}).Debug("required bundle service creation")

		// create the bundle's services on each of the nodes
		for _, nodeIP := range nodeIPs {
			for _, serviceConf := range bundle.Services {
				wg.Add(1)

				trackError := genTrackServiceError(errorChan, &wg, bundle.ID, serviceConf.ID, nodeIP)
				logSuccess := genLogSuccess(bundle.ID, serviceConf.ID, nodeIP)

				opts := acomm.RequestOptions{
					Task:         "service-create",
					ResponseHook: config.HTTPResponseURL(),
					Args: service.CreateArgs{
						ID:       serviceConf.ID,
						BundleID: bundle.ID,
						Dataset:  filepath.Join(config.DatasetPrefix(), serviceConf.Dataset),
						Cmd:      serviceConf.Cmd,
						Env:      serviceConf.Env,
					},
					ErrorHandler: func(req *acomm.Request, resp *acomm.Response) {
						trackError(req.Task, errors.Wrap(resp.Error))
					},
					SuccessHandler: func(_ *acomm.Request, _ *acomm.Response) {
						logSuccess()
						wg.Done()
					},
				}

				if err := tick.SendNodeRequest(config, tracker, opts, nodeIP); err != nil {
					trackError(opts.Task, err)
				}
			}
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
		err = errors.Newv("not all bundle services started", map[string]interface{}{"errors": errs})
	}

	return err
}

func genTrackServiceError(errorChan chan error, wg *sync.WaitGroup, id uint64, service, node string) func(string, error) {
	return func(task string, err error) {
		errorChan <- errors.Wrapv(err, map[string]interface{}{
			"task":    task,
			"bundle":  id,
			"service": service,
			"node":    node,
		})

		wg.Done()
	}
}

func genLogSuccess(id uint64, service, destination string) func() {
	return func() {
		logrus.WithFields(logrus.Fields{
			"bundle":  id,
			"service": service,
			"node":    destination,
		}).Debug("bundle running")
	}
}
