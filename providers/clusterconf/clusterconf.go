package clusterconf

import (
	"encoding/json"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/kv"
	"github.com/cerana/cerana/provider"
)

// ClusterConf is a provider of cluster configuration functionality.
type ClusterConf struct {
	config  *Config
	tracker *acomm.Tracker
}

// IDArgs are arguments for operations requiring only an ID.
type IDArgs struct {
	ID string `json:"id"`
}

// New creates a new instance of ClusterConf
func New(config *Config, tracker *acomm.Tracker) *ClusterConf {
	return &ClusterConf{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (c *ClusterConf) RegisterTasks(server *provider.Server) {
	server.RegisterTask("get-bundle", c.GetBundle)
	server.RegisterTask("list-bundles", c.ListBundles)
	server.RegisterTask("update-bundle", c.UpdateBundle)
	server.RegisterTask("delete-bundle", c.DeleteBundle)
	server.RegisterTask("bundle-heartbeat", c.BundleHeartbeat)
	server.RegisterTask("list-bundle-heartbeats", c.ListBundleHeartbeats)

	server.RegisterTask("get-dataset", c.GetDataset)
	server.RegisterTask("list-datasets", c.ListDatasets)
	server.RegisterTask("update-dataset", c.UpdateDataset)
	server.RegisterTask("delete-dataset", c.DeleteDataset)
	server.RegisterTask("dataset-heartbeat", c.DatasetHeartbeat)
	server.RegisterTask("list-dataset-heartbeats", c.ListDatasetHeartbeats)

	server.RegisterTask("get-default-options", c.GetDefaults)
	server.RegisterTask("set-default-options", c.UpdateDefaults)

	server.RegisterTask("node-heartbeat", c.NodeHeartbeat)
	server.RegisterTask("get-node", c.GetNode)
	server.RegisterTask("list-nodes", c.ListNodes)
	server.RegisterTask("get-nodes-history", c.GetNodesHistory)

	server.RegisterTask("get-service", c.GetService)
	server.RegisterTask("update-service", c.UpdateService)
	server.RegisterTask("delete-service", c.DeleteService)

	server.RegisterTask("get-dhcp", c.GetDHCP)
	server.RegisterTask("set-dhcp", c.SetDHCP)
}

func (c *ClusterConf) kvReq(task string, args map[string]interface{}) (*acomm.Response, error) {
	respChan := make(chan *acomm.Response, 1)
	defer close(respChan)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		respChan <- resp
	}

	if val, ok := args["value"]; ok {
		valJSON, err := json.Marshal(val)
		if err != nil {
			return nil, err
		}
		args["value"] = string(valJSON)
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           task,
		ResponseHook:   c.tracker.URL(),
		Args:           args,
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return nil, err
	}
	if err := c.tracker.TrackRequest(req, 0); err != nil {
		return nil, err
	}
	if err := acomm.Send(c.config.CoordinatorURL(), req); err != nil {
		c.tracker.RemoveRequest(req)
		return nil, err
	}

	resp := <-respChan
	return resp, resp.Error
}

func (c *ClusterConf) kvKeys(prefix string) ([]string, error) {
	args := map[string]interface{}{
		"key": prefix,
	}

	resp, err := c.kvReq("kv-keys", args)
	if err != nil {
		return nil, err
	}
	var values []string
	if err := resp.UnmarshalResult(&values); err != nil {
		return nil, err
	}
	return values, nil
}

func (c *ClusterConf) kvGetAll(key string) (map[string]kv.Value, error) {
	args := map[string]interface{}{
		"key": key,
	}

	resp, err := c.kvReq("kv-getAll", args)
	if err != nil {
		return nil, err
	}
	values := make(map[string]kv.Value)
	if err := resp.UnmarshalResult(&values); err != nil {
		return nil, err
	}
	return values, nil
}

func (c *ClusterConf) kvGet(key string) (kv.Value, error) {
	var value kv.Value
	args := map[string]interface{}{
		"key": key,
	}

	resp, err := c.kvReq("kv-get", args)
	if err != nil {
		return value, err
	}

	if err := resp.UnmarshalResult(&value); err != nil {
		return value, err
	}
	return value, nil
}

func (c *ClusterConf) kvDelete(key string, modIndex uint64) error {
	args := map[string]interface{}{
		"key":     key,
		"recurse": true,
	}
	_, err := c.kvReq("kv-delete", args)
	return err
}

func (c *ClusterConf) kvUpdate(key string, value interface{}, modIndex uint64) (uint64, error) {
	args := map[string]interface{}{
		"key":   key,
		"value": value,
		"index": modIndex,
	}
	resp, err := c.kvReq("kv-update", args)
	if err != nil {
		return 0, err
	}
	result := make(map[string]uint64)
	if err := resp.UnmarshalResult(&result); err != nil {
		return 0, err
	}
	return result["index"], nil
}

func (c *ClusterConf) kvEphemeral(key string, value interface{}, ttl time.Duration) error {
	args := map[string]interface{}{
		"key":   key,
		"value": value,
		"ttl":   ttl,
	}
	_, err := c.kvReq("kv-ephemeral-set", args)
	return err
}
