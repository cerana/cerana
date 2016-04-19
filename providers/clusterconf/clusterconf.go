package clusterconf

import (
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/mistifyio/lochness/pkg/kv"
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
}

func (c *ClusterConf) kvReq(task string, args map[string]interface{}) (*acomm.Response, error) {
	respChan := make(chan *acomm.Response, 1)
	defer close(respChan)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		respChan <- resp
	}

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           task,
		ResponseHook:   c.tracker.URL(),
		Args:           args,
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err := c.tracker.TrackRequest(req, 0); err != nil {
		return nil, err
	}
	if err := acomm.Send(c.config.CoordinatorURL(), req); err != nil {
		c.tracker.RemoveRequest(req)
		return nil, err
	}

	return <-respChan, err
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
	if err := resp.UnmarshalResult(values); err != nil {
		return nil, err
	}
	return values, nil
}

func (c *ClusterConf) kvGet(key string) (*kv.Value, error) {
	args := map[string]interface{}{
		"key": key,
	}

	resp, err := c.kvReq("kv-get", args)
	if err != nil {
		return nil, err
	}

	var value kv.Value
	if err := resp.UnmarshalResult(&value); err != nil {
		return nil, err
	}
	return &value, nil
}

func (c *ClusterConf) kvDelete(key string, modIndex uint64) error {
	args := map[string]interface{}{
		"key":     key,
		"recurse": true,
	}
	resp, err := c.kvReq("kv-delete", args)
	if err != nil {
		return err
	}
	return resp.Error
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
	if err := resp.UnmarshalResult(result); err != nil {
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
	resp, err := c.kvReq("kv-ephemeral", args)
	if err != nil {
		return err
	}
	return resp.Error
}
