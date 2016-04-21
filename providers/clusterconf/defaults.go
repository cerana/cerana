package clusterconf

import (
	"encoding/json"
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
)

const defaultsPrefix string = "cluster"

// Defaults is information about the cluster configuration.
type Defaults struct {
	*DefaultsConf
	c        *ClusterConf
	ModIndex uint64 `json:"modIndex"`
}

// DefaultsConf is the configuration for the cluster.
type DefaultsConf struct {
	ZFSManual bool `json:"zfsManual"`
}

// DefaultsPayload can be used for task args or result when a cluster object
// needs to be sent.
type DefaultsPayload struct {
	Defaults *Defaults `json:"defaults"`
}

// GetDefaults retrieves the cluster config.
func (c *ClusterConf) GetDefaults(req *acomm.Request) (interface{}, *url.URL, error) {
	defaults, err := c.getDefaults()
	if err != nil {
		return nil, nil, err
	}
	return &DefaultsPayload{defaults}, nil, nil
}

// UpdateDefaults sets or updates
func (c *ClusterConf) UpdateDefaults(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DefaultsPayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Defaults == nil {
		return nil, nil, errors.New("missing arg: defaults")
	}

	args.Defaults.c = c

	if err := args.Defaults.update(); err != nil {
		return nil, nil, err
	}
	return &DefaultsPayload{args.Defaults}, nil, nil
}

func (c *ClusterConf) getDefaults() (*Defaults, error) {
	defaults := &Defaults{
		c: c,
	}
	if err := defaults.reload(); err != nil {
		return nil, err
	}
	return defaults, nil
}

func (d *Defaults) reload() error {
	value, err := d.c.kvGet(defaultsPrefix)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(value.Data, d.DefaultsConf); err != nil {
		return err
	}
	d.ModIndex = value.Index
	return nil
}

func (d *Defaults) update() error {
	modIndex, err := d.c.kvUpdate(defaultsPrefix, d.DefaultsConf, d.ModIndex)
	if err != nil {
		return err
	}
	d.ModIndex = modIndex

	return nil
}
