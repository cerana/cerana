package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/spf13/pflag"
)

type demotick struct {
	coordinatorURL      *url.URL
	datasetDir          string
	nodeCoordinatorPort uint
	requestTimeout      time.Duration
	tracker             *acomm.Tracker
	responseHook        *url.URL
}

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})
	var urlS, logLevel, datasetDir, responseAddr string
	var requestTimeout time.Duration
	var nodeCoordinatorPort uint

	pflag.StringVarP(&urlS, "coordinator_url", "c", "", "layer2 coordinator url")
	pflag.StringVarP(&datasetDir, "dataset_dir", "d", "data/datasets", "directory containing datasets on nodes")
	pflag.StringVarP(&logLevel, "log_leveL", "l", "info", "log level: debug/info/warn/error/fatal/panic")
	pflag.DurationVarP(&requestTimeout, "request_timeout", "r", 10*time.Second, "request timeout")
	pflag.UintVarP(&nodeCoordinatorPort, "node_coordinator_port", "n", 8080, "node coordinator port")
	pflag.StringVarP(&responseAddr, "response_addr", "r", ":20000", "demotick external response port")
	pflag.Parse()

	dieOnError(logrusx.SetLevel(logLevel))

	if urlS == "" {
		dieOnError(errors.New("missing coordinator_url"))
	}
	if datasetDir == "" {
		dieOnError(errors.New("missing dataset_dir"))
	}

	u, err := url.Parse(urlS)
	dieOnError(err)

	responseHook, err := url.ParseRequestURI(fmt.Sprintf("http://%s/response", responseAddr))
	dieOnError(err)

	tracker, err := acomm.NewTracker("", nil, nil, requestTimeout)
	dieOnError(err)
	dieOnError(tracker.Start())

	d := demotick{
		coordinatorURL:      u,
		datasetDir:          datasetDir,
		nodeCoordinatorPort: nodeCoordinatorPort,
		requestTimeout:      requestTimeout,
		tracker:             tracker,
		responseHook:        responseHook,
	}

	d.startHTTPResponseServer(responseAddr)

	logrus.Info("running cluster tick")
	dieOnError(d.run())
	logrus.Info("completed cluster tick successfully")
}

func (d *demotick) run() error {
	// Get nodes
	logrus.Info("retrieving list of healthy nodes")
	nodes, err := d.getNodes()
	if err != nil {
		return err
	}

	// Get datasets and current node(s)
	logrus.Info("retrieving datasets and their current locations")
	datasets, err := d.getDatasets()
	if err != nil {
		return err
	}
	// ZFS Send/Receive datasets from current node(s) to missing node(s)
	logrus.Info("replicating datasets to cluster nodes")
	if err = d.replicateDatasets(datasets, nodes); err != nil {
		return err
	}
	logrus.Info("replicating datasets successful")

	// Get bundles and current node(s)
	logrus.Info("retrieving bundles and their current locations")
	bundles, bundleHeartbeats, err := d.getBundles()
	if err != nil {
		return err
	}
	// Start bundles on missing node(s)
	logrus.Info("running bundles on cluster nodes")
	if err = d.runBundles(bundles, bundleHeartbeats, nodes); err != nil {
		return err
	}
	logrus.Info("running bundles on cluster nodes successful")
	return nil
}

func (d *demotick) getNodes() ([]clusterconf.Node, error) {
	opts := acomm.RequestOptions{Task: "list-nodes"}
	resp, err := d.tracker.SyncRequest(d.coordinatorURL, opts, d.requestTimeout)
	if err != nil {
		return nil, err
	}
	var result clusterconf.ListNodesResult
	if err := resp.UnmarshalResult(&result); err != nil {
		return nil, err
	}
	return result.Nodes, nil
}

func (d *demotick) getDatasets() (map[string]map[string]clusterconf.DatasetHeartbeat, error) {
	tasks := map[string]interface{}{
		"list-dataset-heartbeats": &clusterconf.DatasetHeartbeatList{},
		"list-datasets":           &clusterconf.DatasetListResult{},
	}
	multirequest := acomm.NewMultiRequest(d.tracker, d.requestTimeout)
	for task := range tasks {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: task})
		if err != nil {
			return nil, err
		}
		if err := multirequest.AddRequest(task, req); err != nil {
			return nil, err
		}
		if err := acomm.Send(d.coordinatorURL, req); err != nil {
			multirequest.RemoveRequest(req)
			return nil, err
		}
	}

	responses := multirequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			return nil, fmt.Errorf("%s : %s", name, resp.Error)
		}
		if err := resp.UnmarshalResult(tasks[name]); err != nil {
			return nil, fmt.Errorf("%s : %s", name, err)
		}
	}

	heartbeats := tasks["list-dataset-heartbeats"].(*clusterconf.DatasetHeartbeatList).Heartbeats
	datasets := tasks["list-datasets"].(*clusterconf.DatasetListResult).Datasets
	datasetMap := make(map[string]struct{})
	for _, dataset := range datasets {
		datasetMap[dataset.ID] = struct{}{}
	}

	// remove any datasets not tracked by clusterconf
	for datasetID := range heartbeats {
		if _, ok := datasetMap[datasetID]; !ok {
			delete(heartbeats, datasetID)
		}
	}
	return heartbeats, nil
}

func (d *demotick) replicateDatasets(datasets map[string]map[string]clusterconf.DatasetHeartbeat, nodes []clusterconf.Node) error {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(datasets)*len(nodes))
	defer close(errorChan)

	for datasetID, heartbeats := range datasets {
		logrus.Debug("replicating dataset " + datasetID)
		replicateN := len(nodes) - len(heartbeats)
		if replicateN == 0 {
			logrus.Debug("no replication needed for dataset " + datasetID)
			// already present on all nodes
			continue
		}

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

		datasetName := filepath.Join(d.datasetDir, datasetID)

		genSuccessHandler := func(name, dIP string, trackErr func(string, error)) func(*acomm.Request, *acomm.Response) {
			return func(_ *acomm.Request, resp *acomm.Response) {
				logrus.WithField("DatasetStreamURL", resp.StreamURL).Info("zfs-stream url")
				opts := acomm.RequestOptions{
					Task:         "zfs-receive",
					ResponseHook: d.responseHook,
					Args:         zfs.CommonArgs{Name: name},
					StreamURL:    resp.StreamURL,
					ErrorHandler: genErrorHandler(trackErr),
					SuccessHandler: func(_ *acomm.Request, _ *acomm.Response) {
						wg.Done()
					},
				}
				d.sendActionRequest(opts, dIP, trackErr)
			}
		}

		for _, destinationIP := range destinationIPs {
			logrus.Debugf("replicating dataset %s from %s to %d", datasetID, sourceIP, destinationIP)
			wg.Add(1)
			trackErr := func(id, sIP, dIP string) func(string, error) {
				return func(task string, e error) {
					errorChan <- fmt.Errorf("%s: dataset %s from %s to %s: %s", task, id, sIP, dIP, e)
					wg.Done()
				}
			}(datasetID, sourceIP, destinationIP)

			opts := acomm.RequestOptions{
				Task:           "zfs-send",
				ResponseHook:   d.responseHook,
				Args:           zfs.CommonArgs{Name: fmt.Sprintf("%s@%s", datasetName, datasetID)},
				ErrorHandler:   genErrorHandler(trackErr),
				SuccessHandler: genSuccessHandler(datasetName, destinationIP, trackErr),
			}
			d.sendActionRequest(opts, sourceIP, trackErr)
		}
	}

	wg.Wait()

	var err error
	for i := len(errorChan); i > 0; i-- {
		e := <-errorChan
		logrus.WithField("error", e).Error("failed to replicate dataset")
		if err != nil {
			err = errors.New("not all datasets replicated")
		}
	}
	return err
}

func (d *demotick) sendActionRequest(opts acomm.RequestOptions, ip string, trackErr func(string, error)) {
	nodeURL, err := url.Parse(fmt.Sprintf("http://%s:%d", ip, d.nodeCoordinatorPort))
	if err != nil {
		trackErr(opts.Task, err)
		return
	}

	req, err := acomm.NewRequest(opts)
	if err != nil {
		trackErr(req.Task, err)
		return
	}
	if err = d.tracker.TrackRequest(req, d.requestTimeout); err != nil {
		trackErr(req.Task, err)
		return
	}

	if err = acomm.Send(nodeURL, req); err != nil {
		d.tracker.RemoveRequest(req)
		trackErr(req.Task, err)
		return
	}
}

func (d *demotick) getBundles() ([]*clusterconf.Bundle, map[uint64]clusterconf.BundleHeartbeats, error) {
	tasks := map[string]struct {
		args   interface{}
		result interface{}
	}{
		"list-bundle-heartbeats": {nil, &clusterconf.BundleHeartbeatList{}},
		"list-bundles":           {clusterconf.ListBundleArgs{CombinedOverlay: true}, &clusterconf.BundleListResult{}},
	}
	multirequest := acomm.NewMultiRequest(d.tracker, d.requestTimeout)
	for taskName, task := range tasks {
		req, err := acomm.NewRequest(acomm.RequestOptions{Task: taskName, Args: task.args})
		if err != nil {
			return nil, nil, err
		}
		if err := multirequest.AddRequest(taskName, req); err != nil {
			return nil, nil, err
		}
		if err := acomm.Send(d.coordinatorURL, req); err != nil {
			multirequest.RemoveRequest(req)
			return nil, nil, err
		}
	}

	responses := multirequest.Responses()
	for name, resp := range responses {
		if resp.Error != nil {
			return nil, nil, fmt.Errorf("%s : %s", name, resp.Error)
		}
		if err := resp.UnmarshalResult(tasks[name].result); err != nil {
			return nil, nil, fmt.Errorf("%s : %s", name, err)
		}
	}

	heartbeats := tasks["list-bundle-heartbeats"].result.(*clusterconf.BundleHeartbeatList).Heartbeats
	bundles := tasks["list-bundles"].result.(*clusterconf.BundleListResult).Bundles
	bundleMap := make(map[uint64]struct{})
	for _, bundle := range bundles {
		bundleMap[bundle.ID] = struct{}{}
	}

	// remove any bundles not tracked by clusterconf
	for bundleID := range heartbeats {
		if _, ok := bundleMap[bundleID]; !ok {
			delete(heartbeats, bundleID)
		}
	}
	return bundles, heartbeats, nil
}

func (d *demotick) runBundles(bundles []*clusterconf.Bundle, bundleHeartbeats map[uint64]clusterconf.BundleHeartbeats, nodes []clusterconf.Node) error {
	var wg sync.WaitGroup
	// make error channel large enough to handle every service of every bundle on all nodes
	count := 0
	for _, b := range bundles {
		count += len(b.Services)
	}
	errorChan := make(chan error, count*len(nodes))
	defer close(errorChan)

	for _, bundle := range bundles {
		logrus.Debugf("running bundle %d", bundle.ID)
		heartbeats := bundleHeartbeats[bundle.ID]

		replicateN := len(nodes) - len(heartbeats)
		if replicateN == 0 {
			logrus.Debugf("no running needed for bundle %d", bundle.ID)
			// already present on all nodes
			continue
		}

		destinationIPs := make([]string, 0, replicateN)
		for _, node := range nodes {
			if heartbeats == nil {
				destinationIPs = append(destinationIPs, node.ID)
			} else if _, ok := heartbeats[node.ID]; !ok {
				destinationIPs = append(destinationIPs, node.ID)
			}
		}

		for _, destinationIP := range destinationIPs {
			for _, serviceConf := range bundle.Services {
				logrus.Debugf("running bundle %d service %s on %s", bundle.ID, serviceConf.ID, destinationIP)
				wg.Add(1)
				trackErr := func(bID uint64, sID, dIP string) func(string, error) {
					return func(task string, e error) {
						errorChan <- fmt.Errorf("%s: run service %d:%s on %s: %s", task, bID, sID, dIP, e)
						wg.Done()
					}
				}(bundle.ID, serviceConf.ID, destinationIP)

				opts := acomm.RequestOptions{
					Task:         "service-create",
					ResponseHook: d.responseHook,
					Args: service.CreateArgs{
						ID:       serviceConf.ID,
						BundleID: bundle.ID,
						Dataset:  serviceConf.Dataset,
						Cmd:      serviceConf.Cmd,
						Env:      serviceConf.Env,
					},
					ErrorHandler: genErrorHandler(trackErr),
					SuccessHandler: func(_ *acomm.Request, _ *acomm.Response) {
						wg.Done()
					},
				}
				d.sendActionRequest(opts, destinationIP, trackErr)
			}
		}
	}

	wg.Wait()

	var err error
	for i := len(errorChan); i > 0; i-- {
		e := <-errorChan
		logrus.WithField("error", e).Error("failed to start bundle service")
		if err != nil {
			err = errors.New("not all bundles started")
		}
	}
	return err
}

func genErrorHandler(trackErr func(string, error)) acomm.ResponseHandler {
	return func(req *acomm.Request, resp *acomm.Response) {
		trackErr(req.Task, resp.Error)
	}
}

func (d *demotick) startHTTPResponseServer(addr string) {
	http.HandleFunc("/response", d.tracker.ProxyExternalHandler)
	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil {
			dieOnError(err)
		}
	}()
}

func dieOnError(err error) {
	if err != nil {
		logrus.WithField("error", err).Fatal("encountered an error during startup")
	}
}
