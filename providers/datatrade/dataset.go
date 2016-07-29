package datatrade

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/pborman/uuid"
)

// DatasetImportArgs are arguments for configuring an imported dataset.
type DatasetImportArgs struct {
	NFS        bool   `json:"nfs"`
	Quota      uint64 `json:"quota"`
	ReadOnly   bool   `json:"readOnly"`
	Redundancy uint64 `json:"redundancy"`
}

// DatasetImportResult is the result of a dataset import.
type DatasetImportResult struct {
	Dataset clusterconf.Dataset `json:"dataset"`
	NodeID  string              `json:"nodeID"`
}

// DatasetImport imports a dataset into the cluster and tracks it in the
// cluster configuration.
func (p *Provider) DatasetImport(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DatasetImportArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Redundancy == 0 {
		return nil, nil, errors.New("missing arg: redundancy")
	}

	if req.StreamURL == nil {
		return nil, nil, errors.New("missing request streamURL")
	}

	dataset := clusterconf.Dataset{
		ID:         uuid.New(),
		NFS:        args.ReadOnly,
		Quota:      args.Quota,
		ReadOnly:   args.ReadOnly,
		Redundancy: args.Redundancy,
	}

	logrus.Info("selecting import node for dataset")
	node, err := p.datasetImportNode()
	if err != nil {
		return nil, nil, err
	}

	logrus.Info("importing dataset")
	if err := p.datasetImport(node.ID, dataset.ID, req.StreamURL); err != nil {
		return nil, nil, err
	}

	if args.ReadOnly {
		logrus.Info("taking snapshot of imported ro dataset")
		if err := p.datasetSnapshot(node.ID, dataset.ID); err != nil {
			return nil, nil, err
		}
	}

	return DatasetImportResult{Dataset: dataset, NodeID: node.ID}, nil, p.datasetConfig(dataset)

}

func (p *Provider) datasetImportNode() (*clusterconf.Node, error) {
	opts := acomm.RequestOptions{
		Task: "list-nodes",
	}
	resp, err := p.tracker.SyncRequest(p.config.CoordinatorURL(), opts, p.config.RequestTimeout())
	if err != nil {
		return nil, err
	}
	var result clusterconf.ListNodesResult
	if err := resp.UnmarshalResult(&result); err != nil {
		return nil, err
	}
	if len(result.Nodes) == 0 {
		return nil, errors.New("no nodes found")
	}
	node := result.Nodes[rand.Intn(len(result.Nodes))]
	return &node, nil
}

func (p *Provider) datasetImport(nodeID, datasetID string, streamURL *url.URL) error {
	taskURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%d", nodeID, p.config.NodeCoordinatorPort()))
	if err != nil {
		return err
	}
	opts := acomm.RequestOptions{
		Task:      "zfs-receive",
		TaskURL:   taskURL,
		StreamURL: streamURL,
		Args: zfs.CommonArgs{
			Name: filepath.Join(p.config.DatasetDir(), datasetID),
		},
	}
	logrus.WithField("requestOpts", opts).Info("sending dataset import request to node")
	_, err = p.tracker.SyncRequest(p.config.CoordinatorURL(), opts, p.config.RequestTimeout())
	return err
}

func (p *Provider) datasetSnapshot(nodeID, datasetID string) error {
	taskURL, err := url.ParseRequestURI(fmt.Sprintf("http://%s:%d", nodeID, p.config.NodeCoordinatorPort()))
	if err != nil {
		return err
	}
	opts := acomm.RequestOptions{
		Task:    "zfs-snapshot",
		TaskURL: taskURL,
		Args: zfs.SnapshotArgs{
			Name:      filepath.Join(p.config.DatasetDir(), datasetID),
			SnapName:  datasetID,
			Recursive: false,
		},
	}
	logrus.WithField("requestOpts", opts).Info("snapshotting imported dataset on node")
	_, err = p.tracker.SyncRequest(p.config.CoordinatorURL(), opts, p.config.RequestTimeout())
	return err
}

func (p *Provider) datasetConfig(dataset clusterconf.Dataset) error {
	logrus.Info("updating clusterconf with dataset info")
	opts := acomm.RequestOptions{
		Task: "update-dataset",
		Args: clusterconf.DatasetPayload{Dataset: &dataset},
	}
	_, err := p.tracker.SyncRequest(p.config.CoordinatorURL(), opts, p.config.RequestTimeout())
	logrus.Info("done updating clusterconf with dataset info")
	return err
}
