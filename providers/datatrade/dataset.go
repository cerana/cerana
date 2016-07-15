package datatrade

import (
	"errors"
	"fmt"
	"math/rand"
	"net/url"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/pborman/uuid"
)

// DatasetImportArgs are arguments for configuring an imported dataset.
type DatasetImportArgs struct {
	NFS        bool `json:"nfs"`
	Quota      int  `json:"quota"`
	ReadOnly   bool `json:"readOnly"`
	Redundancy int  `json:"redundancy"`
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

	node, err := p.datasetImportNode()
	if err != nil {
		return nil, nil, err
	}

	if err := p.datasetImport(node.ID, dataset.ID, req.StreamURL); err != nil {
		return nil, nil, err
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
	opts := acomm.RequestOptions{
		Task:      "zfs-receive",
		StreamURL: streamURL,
		Args: zfs.CommonArgs{
			Name: filepath.Join(p.config.DatasetDir(), datasetID),
		},
	}
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", nodeID, p.config.NodeCoordinatorPort()))
	if err != nil {
		return err
	}
	_, err = p.tracker.SyncRequest(u, opts, p.config.RequestTimeout())
	return err
}

func (p *Provider) datasetConfig(dataset clusterconf.Dataset) error {
	opts := acomm.RequestOptions{
		Task: "update-dataset",
		Args: clusterconf.DatasetPayload{Dataset: &dataset},
	}
	_, err := p.tracker.SyncRequest(p.config.CoordinatorURL(), opts, p.config.RequestTimeout())
	return err
}
