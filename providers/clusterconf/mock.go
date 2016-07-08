package clusterconf

import (
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/pborman/uuid"
)

// MockClusterConf is a mock ClusterConf provider.
type MockClusterConf struct {
	Data *MockClusterData
}

// MockClusterData is the in-memory data structure for a MockClusterConf.
type MockClusterData struct {
	Services   map[string]*Service
	Bundles    map[uint64]*Bundle
	BundlesHB  map[uint64]BundleHeartbeats
	Datasets   map[string]*Dataset
	DatasetsHB map[string]map[string]DatasetHeartbeat
	Nodes      map[string]*Node
	History    NodesHistory
	Defaults   *Defaults
}

// NewMockClusterConf creates a new MockClusterConf.
func NewMockClusterConf() *MockClusterConf {
	return &MockClusterConf{
		Data: &MockClusterData{
			Services:   make(map[string]*Service),
			Bundles:    make(map[uint64]*Bundle),
			BundlesHB:  make(map[uint64]BundleHeartbeats),
			Datasets:   make(map[string]*Dataset),
			DatasetsHB: make(map[string]map[string]DatasetHeartbeat),
			Nodes:      make(map[string]*Node),
			History:    make(NodesHistory),
		},
	}
}

// RegisterTasks registers all of MockClusterConf's tasks.
func (c *MockClusterConf) RegisterTasks(server *provider.Server) {
	server.RegisterTask("get-bundle", c.GetBundle)
	server.RegisterTask("list-bundles", c.ListBundles)
	server.RegisterTask("list-bundle-heartbeats", c.ListBundleHeartbeats)
	server.RegisterTask("update-bundle", c.UpdateBundle)
	server.RegisterTask("delete-bundle", c.DeleteBundle)
	server.RegisterTask("bundle-heartbeat", c.BundleHeartbeat)
	server.RegisterTask("get-dataset", c.GetDataset)
	server.RegisterTask("list-datasets", c.ListDatasets)
	server.RegisterTask("list-dataset-heartbeats", c.ListDatasetHeartbeats)
	server.RegisterTask("update-dataset", c.UpdateDataset)
	server.RegisterTask("delete-dataset", c.DeleteDataset)
	server.RegisterTask("dataset-heartbeat", c.DatasetHeartbeat)
	server.RegisterTask("get-default-options", c.GetDefaults)
	server.RegisterTask("set-default-options", c.UpdateDefaults)
	server.RegisterTask("node-heartbeat", c.NodeHeartbeat)
	server.RegisterTask("get-node", c.GetNode)
	server.RegisterTask("get-nodes-history", c.GetNodesHistory)
	server.RegisterTask("get-service", c.GetService)
	server.RegisterTask("update-service", c.UpdateService)
	server.RegisterTask("delete-service", c.DeleteService)
}

// GetBundle retrieves a mock bundle.
func (c *MockClusterConf) GetBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetBundleArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}
	bundle, ok := c.Data.Bundles[args.ID]
	if !ok {
		return nil, nil, errors.New("bundle config not found")
	}
	return &BundlePayload{bundle}, nil, nil
}

// ListBundles retrieves all mock bundles.
func (c *MockClusterConf) ListBundles(req *acomm.Request) (interface{}, *url.URL, error) {
	bundles := make([]*Bundle, 0, len(c.Data.Bundles))
	for _, bundle := range c.Data.Bundles {
		bundles = append(bundles, bundle)
	}
	return &BundleListResult{bundles}, nil, nil
}

// UpdateBundle updates a mock bundle.
func (c *MockClusterConf) UpdateBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundlePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Bundle == nil {
		return nil, nil, errors.New("missing arg: bundle")
	}

	if args.Bundle.ID == 0 {
		rand.Seed(time.Now().UnixNano())
		args.Bundle.ID = uint64(rand.Int63())
	}

	args.Bundle.ModIndex++
	c.Data.Bundles[args.Bundle.ID] = args.Bundle
	return &BundlePayload{args.Bundle}, nil, nil
}

// DeleteBundle removes a mock bundle.
func (c *MockClusterConf) DeleteBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DeleteBundleArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}

	delete(c.Data.Bundles, args.ID)
	return nil, nil, nil
}

// BundleHeartbeat adds a mock bundle heartbeat.
func (c *MockClusterConf) BundleHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleHeartbeatArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.Serial == "" {
		return nil, nil, errors.New("missing arg: serial")
	}
	if args.IP == nil {
		return nil, nil, errors.New("missing arg: ip")
	}

	if _, ok := c.Data.BundlesHB[args.ID]; !ok {
		c.Data.BundlesHB[args.ID] = make(BundleHeartbeats)
	}
	c.Data.BundlesHB[args.ID][args.Serial] = BundleHeartbeat{IP: args.IP, HealthErrors: args.HealthErrors}
	return nil, nil, nil
}

// ListBundleHeartbeats list all mock bundle heartbeats.
func (c *MockClusterConf) ListBundleHeartbeats(req *acomm.Request) (interface{}, *url.URL, error) {
	return BundleHeartbeatList{c.Data.BundlesHB}, nil, nil
}

// GetDataset retrieves a mock dataset.
func (c *MockClusterConf) GetDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	dataset, ok := c.Data.Datasets[args.ID]
	if !ok {
		return nil, nil, errors.New("dataset config not found")
	}
	return &DatasetPayload{dataset}, nil, nil
}

// ListDatasets lists all mock datasets.
func (c *MockClusterConf) ListDatasets(req *acomm.Request) (interface{}, *url.URL, error) {
	datasets := make([]*Dataset, 0, len(c.Data.Datasets))
	for _, dataset := range c.Data.Datasets {
		datasets = append(datasets, dataset)
	}
	return &DatasetListResult{datasets}, nil, nil
}

// UpdateDataset updates a mock dataset.
func (c *MockClusterConf) UpdateDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DatasetPayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Dataset == nil {
		return nil, nil, errors.New("missing arg: dataset")
	}

	if args.Dataset.ID == "" {
		args.Dataset.ID = uuid.New()
	}

	args.Dataset.ModIndex++
	c.Data.Datasets[args.Dataset.ID] = args.Dataset
	return &DatasetPayload{args.Dataset}, nil, nil
}

// DeleteDataset removes a mock dataset.
func (c *MockClusterConf) DeleteDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	delete(c.Data.Datasets, args.ID)
	return nil, nil, nil
}

// DatasetHeartbeat adds a mock dataset heartbeat.
func (c *MockClusterConf) DatasetHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DatasetHeartbeatArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.IP == nil {
		return nil, nil, errors.New("missing arg: ip")
	}
	if _, ok := c.Data.DatasetsHB[args.ID]; !ok {
		c.Data.DatasetsHB[args.ID] = make(map[string]DatasetHeartbeat)
	}
	c.Data.DatasetsHB[args.ID][args.IP.String()] = DatasetHeartbeat{IP: args.IP, InUse: args.InUse}
	return nil, nil, nil
}

// ListDatasetHeartbeats lists all mock dataset heartbeats.
func (c *MockClusterConf) ListDatasetHeartbeats(req *acomm.Request) (interface{}, *url.URL, error) {
	return DatasetHeartbeatList{c.Data.DatasetsHB}, nil, nil
}

// GetDefaults retrieves the mock default values.
func (c *MockClusterConf) GetDefaults(req *acomm.Request) (interface{}, *url.URL, error) {
	return &DefaultsPayload{c.Data.Defaults}, nil, nil
}

// UpdateDefaults updates the mock default values.
func (c *MockClusterConf) UpdateDefaults(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DefaultsPayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Defaults == nil {
		return nil, nil, errors.New("missing arg: defaults")
	}

	args.Defaults.ModIndex++
	c.Data.Defaults = args.Defaults
	return &DefaultsPayload{args.Defaults}, nil, nil
}

// NodeHeartbeat adds a mock node heartbeat.
func (c *MockClusterConf) NodeHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
	var args NodePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Node == nil {
		return nil, nil, errors.New("missing arg: node")
	}

	c.Data.Nodes[args.Node.ID] = args.Node
	history, ok := c.Data.History[args.Node.ID]
	if ok {
		history[args.Node.Heartbeat] = args.Node
	} else {
		c.Data.History[args.Node.ID] = NodeHistory{
			args.Node.Heartbeat: args.Node,
		}
	}
	return nil, nil, nil
}

// GetNode retrieves a mock node.
func (c *MockClusterConf) GetNode(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	node, ok := c.Data.Nodes[args.ID]
	if !ok {
		return nil, nil, errors.New("node not found")
	}
	return &NodePayload{node}, nil, nil
}

// GetNodesHistory retrieves mock nodes history.
func (c *MockClusterConf) GetNodesHistory(req *acomm.Request) (interface{}, *url.URL, error) {
	var args NodeHistoryArgs
	if err := req.UnmarshalArgs(args); err != nil {
		return nil, nil, err
	}

	filters := []nodeFilter{
		nodeFilterID(args.IDs...),
		nodeFilterHeartbeat(args.Before, args.After),
	}

	history := make(NodesHistory)
	for _, savedNodeHistory := range c.Data.History {
		for _, node := range savedNodeHistory {
			for _, fn := range filters {
				if !fn(*node) {
					continue
				}
			}

			nodeHistory, ok := history[node.ID]
			if !ok {
				nodeHistory = make(NodeHistory)
				history[node.ID] = nodeHistory
			}
			nodeHistory[node.Heartbeat] = node
		}
	}

	return &NodesHistoryResult{history}, nil, nil
}

// GetService retrieves a mock service.
func (c *MockClusterConf) GetService(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	service, ok := c.Data.Services[args.ID]
	if !ok {
		return nil, nil, errors.New("service config not found")
	}
	return &ServicePayload{service}, nil, nil
}

// UpdateService updates a mock service.
func (c *MockClusterConf) UpdateService(req *acomm.Request) (interface{}, *url.URL, error) {
	var args ServicePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Service == nil {
		return nil, nil, errors.New("missing arg: service")
	}

	if args.Service.ID == "" {
		args.Service.ID = uuid.New()
	}

	args.Service.ModIndex++
	c.Data.Services[args.Service.ID] = args.Service
	return &ServicePayload{args.Service}, nil, nil
}

// DeleteService removes a mock service.
func (c *MockClusterConf) DeleteService(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	delete(c.Data.Services, args.ID)
	return nil, nil, nil
}
