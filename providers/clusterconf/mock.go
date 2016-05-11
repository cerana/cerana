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

type MockClusterConf struct {
	config *provider.Config
	Data   *MockClusterData
}

type MockClusterData struct {
	Services map[string]*Service
	Bundles  map[int]*Bundle
	Datasets map[string]*Dataset
	Nodes    map[string]*Node
	History  NodesHistory
	Defaults *Defaults
}

func NewMockClusterConf(config *provider.Config) *MockClusterConf {
	return &MockClusterConf{
		config: config,
		Data: &MockClusterData{
			Services: make(map[string]*Service),
			Bundles:  make(map[int]*Bundle),
			Datasets: make(map[string]*Dataset),
			Nodes:    make(map[string]*Node),
			History:  make(NodesHistory),
		},
	}
}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (c *MockClusterConf) RegisterTasks(server *provider.Server) {
	server.RegisterTask("get-bundle", c.GetBundle)
	server.RegisterTask("list-bundles", c.ListBundles)
	server.RegisterTask("update-bundle", c.UpdateBundle)
	server.RegisterTask("delete-bundle", c.DeleteBundle)
	server.RegisterTask("bundle-heartbeat", c.BundleHeartbeat)
	server.RegisterTask("get-dataset", c.GetDataset)
	server.RegisterTask("list-datasets", c.ListDatasets)
	server.RegisterTask("update-dataset", c.UpdateDataset)
	server.RegisterTask("delete-dataset", c.DeleteDataset)
	server.RegisterTask("dataset-heartbeat", c.DeleteDataset)
	server.RegisterTask("get-default-options", c.GetDefaults)
	server.RegisterTask("set-default-options", c.UpdateDefaults)
	server.RegisterTask("node-heartbeat", c.NodeHeartbeat)
	server.RegisterTask("get-node", c.GetNode)
	server.RegisterTask("get-nodes-history", c.GetNodesHistory)
	server.RegisterTask("get-service", c.GetService)
	server.RegisterTask("update-service", c.UpdateService)
	server.RegisterTask("delete-service", c.DeleteService)
}

func (c *MockClusterConf) GetBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleIDArgs
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

func (c *MockClusterConf) ListBundles(req *acomm.Request) (interface{}, *url.URL, error) {
	bundles := make([]*Bundle, 0, len(c.Data.Bundles))
	for _, bundle := range c.Data.Bundles {
		bundles = append(bundles, bundle)
	}
	return &BundleListResult{bundles}, nil, nil
}

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
		args.Bundle.ID = rand.Int()
	}

	args.Bundle.ModIndex++
	c.Data.Bundles[args.Bundle.ID] = args.Bundle
	return &BundlePayload{args.Bundle}, nil, nil
}

func (c *MockClusterConf) DeleteBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleIDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}

	_, ok := c.Data.Bundles[args.ID]
	if !ok {
		return nil, nil, errors.New("bundle config not found")
	}

	delete(c.Data.Bundles, args.ID)
	return nil, nil, nil
}

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

	bundle, ok := c.Data.Bundles[args.ID]
	if !ok {
		return nil, nil, errors.New("bundle config not found")
	}

	bundle.Nodes[args.Serial] = args.IP
	return &BundlePayload{bundle}, nil, nil
}

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
func (c *MockClusterConf) ListDatasets(req *acomm.Request) (interface{}, *url.URL, error) {
	datasets := make([]*Dataset, 0, len(c.Data.Datasets))
	for _, dataset := range c.Data.Datasets {
		datasets = append(datasets, dataset)
	}
	return &DatasetListResult{datasets}, nil, nil
}
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
func (c *MockClusterConf) DeleteDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	_, ok := c.Data.Datasets[args.ID]
	if !ok {
		return nil, nil, errors.New("dataset config not found")
	}

	delete(c.Data.Datasets, args.ID)
	return nil, nil, nil
}
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
	dataset, ok := c.Data.Datasets[args.ID]
	if !ok {
		return nil, nil, errors.New("dataset config not found")
	}
	dataset.Nodes[args.IP.String()] = true
	return &DatasetPayload{dataset}, nil, nil
}

func (c *MockClusterConf) GetDefaults(req *acomm.Request) (interface{}, *url.URL, error) {
	return &DefaultsPayload{c.Data.Defaults}, nil, nil
}

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
func (c *MockClusterConf) DeleteService(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	_, ok := c.Data.Services[args.ID]
	if !ok {
		return nil, nil, errors.New("service config not found")
	}

	delete(c.Data.Services, args.ID)
	return nil, nil, nil
}
