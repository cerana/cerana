package clusterconf

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"path"
	"path/filepath"

	"github.com/cerana/cerana/acomm"
	"github.com/pborman/uuid"
)

const datasetsPrefix string = "datasets"

// Dataset is information about a dataset.
type Dataset struct {
	*DatasetConf
	c *ClusterConf
	// Nodes contains the set of nodes on which the dataset is currently in use.
	// The map keys are IP address strings.
	Nodes map[string]bool `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}

// DatasetConf is the configuration of a dataset.
type DatasetConf struct {
	ID                string `json:"id"`
	Parent            string `json:"parent"`
	ParentSameMachine bool   `json:"parentSameMachine"`
	ReadOnly          bool   `json:"readOnly"`
	NFS               bool   `json:"nfs"`
	Redundancy        int    `json:"redundancy"`
	Quota             int    `json:"quota"`
}

// DatasetPayload can be used for task args or result when a dataset object
// needs to be sent.
type DatasetPayload struct {
	Dataset *Dataset `json:"dataset"`
}

// DatasetHeartbeatArgs are arguments for updating a dataset node heartbeat.
type DatasetHeartbeatArgs struct {
	ID string `json:"id"`
	IP net.IP `json:"ip"`
}

// GetDataset retrieves a dataset.
func (c *ClusterConf) GetDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	dataset, err := c.getDataset(args.ID)
	if err != nil {
		return nil, nil, err
	}
	return &DatasetPayload{dataset}, nil, nil
}

// UpdateDataset creates or updates a dataset config. When updating, a Get should first be performed and the modified Dataset passed back.
func (c *ClusterConf) UpdateDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DatasetPayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Dataset == nil {
		return nil, nil, errors.New("missing arg: dataset")
	}
	args.Dataset.c = c

	if args.Dataset.ID == "" {
		args.Dataset.ID = uuid.New()
	}

	if err := args.Dataset.update(); err != nil {
		return nil, nil, err
	}
	return &DatasetPayload{args.Dataset}, nil, nil
}

// DeleteDataset deletes a dataset config.
func (c *ClusterConf) DeleteDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	dataset, err := c.getDataset(args.ID)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, dataset.delete()
}

// DatasetHeartbeat registers a new node heartbeat that is using the dataset.
func (c *ClusterConf) DatasetHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
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

	dataset, err := c.getDataset(args.ID)
	if err != nil {
		return nil, nil, err
	}

	if err := dataset.nodeHeartbeat(args.IP); err != nil {
		return nil, nil, err
	}

	return &DatasetPayload{dataset}, nil, nil
}

func (c *ClusterConf) getDataset(id string) (*Dataset, error) {
	dataset := &Dataset{
		c:           c,
		DatasetConf: &DatasetConf{ID: id},
	}
	if err := dataset.reload(); err != nil {
		return nil, err
	}
	return dataset, nil
}

func (d *Dataset) reload() error {
	var err error
	key := path.Join(datasetsPrefix, d.ID)
	values, err := d.c.kvGetAll(key) // Blocking
	if err != nil {
		return err
	}

	// Config
	config, ok := values[path.Join(key, "config")]
	if !ok {
		return errors.New("dataset config not found")
	}
	if err = json.Unmarshal(config.Data, d.DatasetConf); err != nil {
		return err
	}
	d.ModIndex = config.Index

	// Nodes
	d.Nodes = make(map[string]bool)
	for key := range values {
		base := filepath.Base(key)
		dir := filepath.Base(filepath.Dir(key))
		if dir == "nodes" {
			d.Nodes[base] = true
		}
	}

	return nil
}

func (d *Dataset) delete() error {
	key := path.Join(datasetsPrefix, d.ID)
	return d.c.kvDelete(key, d.ModIndex)
}

// update saves the core dataset config. It will not modify nodes.
func (d *Dataset) update() error {
	key := path.Join(datasetsPrefix, d.ID, "config")

	_, err := d.c.kvUpdate(key, d.DatasetConf, d.ModIndex)
	if err != nil {
		return err
	}

	// reload instead of just setting the new modIndex in case any nodes have also changed.
	return d.reload()
}

func (d *Dataset) nodeHeartbeat(ip net.IP) error {
	key := path.Join(datasetsPrefix, d.ID, "nodes", ip.String())
	if err := d.c.kvEphemeral(key, true, d.c.config.DatasetTTL()); err != nil {
		return err
	}
	return d.reload()
}
