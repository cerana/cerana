package clusterconf

import (
	"encoding/json"
	"errors"
	"math/rand"
	"net"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cerana/cerana/acomm"
)

const bundlesPrefix string = "bundles"

// Bundle is information about a bundle of services.
type Bundle struct {
	*BundleConf
	c *ClusterConf
	// Nodes contains the set of nodes on which the dataset is currently in use.
	// THe map keys are serials.
	Nodes map[string]net.IP `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}

// BundleConf is the configuration of a bundle
type BundleConf struct {
	ID         int                       `json:"id"`
	Bundles    map[string]*BundleBundle  `json:"bundles"`
	Services   map[string]*BundleService `json:"services"`
	Redundancy int                       `json:"redundancy"`
	Ports      map[int]*BundlePort       `json:"ports"`
}

// BundleBundle is configuration for a bundle associated with a bundle.
type BundleBundle struct {
	Name     string `json:"name"`
	BundleID string `json:"bundleID"`
	Type     int    `json:"type"` // TODO: Decide on type for this. Iota?
	Quota    int    `json:"type"`
}

// BundleService is configuration overrides for a service of a bundle and
// associated bundles.
type BundleService struct {
	*ServiceConf
	Bundles map[string]*ServiceBundle `json:"bundles"`
}

// ServiceBundle is configuration for mounting a bundle for a bundle service.
type ServiceBundle struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountPoint"`
	ReadOnly   bool   `json:"readOnly"`
}

// BundlePort is configuration for a port associated with a bundle.
type BundlePort struct {
	Port             int      `json:"port"`
	Public           bool     `json:"public"`
	ConnectedBundles []string `json:"connectedBundles"`
	ExternalPort     int      `json:"externalPort"`
}

// BundleIDArgs are args for bundle tasks that only require bundle id.
type BundleIDArgs struct {
	ID int `json:"id"`
}

// BundlePayload can be used for task args or result when a bundle object needs
// to be sent.
type BundlePayload struct {
	Bundle *Bundle `json:"bundle"`
}

// BundleHeartbeatArgs are arguments for updating a dataset node heartbeat.
type BundleHeartbeatArgs struct {
	ID     int    `json:"id"`
	Serial string `json:"serial"`
	IP     net.IP `json:"ip"`
}

// GetBundle retrieves a bundle.
func (c *ClusterConf) GetBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleIDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}

	bundle, err := c.getBundle(args.ID)
	if err != nil {
		return nil, nil, err
	}
	return &BundlePayload{bundle}, nil, nil
}

// UpdateBundle creates or updates a bundle config. When updating, a Get should first be performed and the modified Bundle passed back.
func (c *ClusterConf) UpdateBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundlePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Bundle != nil {
		return nil, nil, errors.New("missing arg: bundle")
	}
	args.Bundle.c = c

	if args.Bundle.ID == 0 {
		rand.Seed(time.Now().UnixNano())
		args.Bundle.ID = rand.Int()
	}

	if err := args.Bundle.update(); err != nil {
		return nil, nil, err
	}
	return &BundlePayload{args.Bundle}, nil, nil
}

// DeleteBundle deletes a bundle config.
func (c *ClusterConf) DeleteBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleIDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}

	bundle, err := c.getBundle(args.ID)
	if err != nil {
		return nil, nil, err
	}

	return nil, nil, bundle.delete()
}

// BundleHeartbeat registers a new node heartbeat that is using the bundle.
func (c *ClusterConf) BundleHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundleHeartbeatArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: ID")
	}
	if args.IP == nil {
		return nil, nil, errors.New("missing arg: IP")
	}

	bundle, err := c.getBundle(args.ID)
	if err != nil {
		return nil, nil, err
	}

	if err := bundle.nodeHeartbeat(args.IP); err != nil {
		return nil, nil, err
	}

	return &BundlePayload{bundle}, nil, nil
}

func (c *ClusterConf) getBundle(id int) (*Bundle, error) {
	bundle := &Bundle{
		c:          c,
		BundleConf: &BundleConf{ID: id},
	}
	if err := bundle.reload(); err != nil {
		return nil, err
	}
	return bundle, nil
}

func (b *Bundle) reload() error {
	var err error
	key := path.Join(bundlesPrefix, strconv.Itoa(b.ID))
	values, err := b.c.kvGetAll(key) // Blocking
	if err != nil {
		return err
	}

	// Config
	config, ok := values[path.Join(key, "config")]
	if !ok {
		return errors.New("bundle config not found")
	}
	if err = json.Unmarshal(config.Data, b.BundleConf); err != nil {
		return err
	}
	b.ModIndex = config.Index

	// Nodes
	b.Nodes = make(map[string]net.IP)
	for key, value := range values {
		base := filepath.Base(key)
		dir := filepath.Base(filepath.Dir(key))
		if dir == "nodes" {
			var ip net.IP
			if err := json.Unmarshal(value.Data, &ip); err != nil {
				return err
			}
			b.Nodes[base] = ip
		}
	}

	return nil
}

func (b *Bundle) delete() error {
	key := path.Join(bundlesPrefix, strconv.Itoa(b.ID))
	return b.c.kvDelete(key, b.ModIndex)
}

// update saves the core bundle config. It will not modify nodes.
func (b *Bundle) update() error {
	key := path.Join(bundlesPrefix, strconv.Itoa(b.ID), "config")

	_, err := b.c.kvUpdate(key, b.BundleConf, b.ModIndex)
	if err != nil {
		return err
	}

	// reload instead of just setting the new modIndex in case any nodes have also changed.
	return b.reload()
}

func (b *Bundle) nodeHeartbeat(ip net.IP) error {
	key := path.Join(bundlesPrefix, strconv.Itoa(b.ID), "nodes", ip.String())
	if err := b.c.kvEphemeral(key, true, b.c.config.BundleTTL()); err != nil {
		return err
	}
	return b.reload()
}
