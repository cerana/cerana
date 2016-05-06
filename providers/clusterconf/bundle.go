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
	"strings"
	"sync"
	"time"

	"github.com/cerana/cerana/acomm"
)

const bundlesPrefix string = "bundles"

// Bundle is information about a bundle of services.
type Bundle struct {
	*BundleConf
	c *ClusterConf
	// Nodes contains the set of nodes on which the dataset is currently in use.
	// The map keys are serials.
	Nodes map[string]net.IP `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}

// BundleConf is the configuration of a bundle.
type BundleConf struct {
	ID         uint64                    `json:"id"`
	Datasets   map[string]*BundleDataset `json:"datasets"`
	Services   map[string]*BundleService `json:"services"`
	Redundancy int                       `json:"redundancy"`
	Ports      BundlePorts               `json:"ports"`
}

// BundlePorts is a map of port numbers to port information.
type BundlePorts map[int]*BundlePort

// MarshalJSON marshals BundlePorts into a JSON map, converting int keys to
// strings.
func (p BundlePorts) MarshalJSON() ([]byte, error) {
	ports := make(map[string]*BundlePort)
	for port, value := range p {
		ports[strconv.Itoa(port)] = value
	}
	return json.Marshal(ports)
}

// UnmarshalJSON unmarshals JSON into a BundlePorts, converting string keys to
// ints.
func (p BundlePorts) UnmarshalJSON(data []byte) error {
	ports := make(map[string]*BundlePort)
	if err := json.Unmarshal(data, &ports); err != nil {
		return err
	}

	p = make(BundlePorts)
	for port, value := range ports {
		portI, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		p[portI] = value
	}
	return nil
}

// BundleDataset is configuration for a dataset associated with a bundle.
type BundleDataset struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Type  int    `json:"type"` // TODO: Decide on type for this. Iota?
	Quota int    `json:"type"`
}

// BundleService is configuration overrides for a service of a bundle and
// associated bundles.
type BundleService struct {
	*ServiceConf
	Datasets map[string]*ServiceDataset `json:"datasets"`
}

// ServiceDataset is configuration for mounting a dataset for a bundle service.
type ServiceDataset struct {
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
	ID uint64 `json:"id"`
}

// BundlePayload can be used for task args or result when a bundle object needs
// to be sent.
type BundlePayload struct {
	Bundle *Bundle `json:"bundle"`
}

// BundleListResult is the result from listing bundles.
type BundleListResult struct {
	Bundles []*Bundle `json:"bundles"`
}

// BundleHeartbeatArgs are arguments for updating a dataset node heartbeat.
type BundleHeartbeatArgs struct {
	ID     uint64 `json:"id"`
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

// ListBundle retrieves a bundle.
func (c *ClusterConf) ListBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	keys, err := c.kvKeys(bundlesPrefix)
	if err != nil {
		return nil, nil, err
	}
	// extract and deduplicate the bundle ids
	ids := make(map[int]bool)
	for _, key := range keys {
		// keys are full paths and include all child keys.
		// e.g. {prefix}/{id}/{rest/of/path}
		idS := strings.Split(strings.TrimPrefix(key, bundlesPrefix), "/")[0]
		id, err := strconv.Atoi(idS)
		if err != nil {
			return nil, nil, errors.New("invalid bundle id")
		}
		ids[id] = true
	}

	var wg sync.WaitGroup
	dsChan := make(chan *Bundle, len(ids))
	defer close(dsChan)
	errChan := make(chan error, len(ids))
	defer close(errChan)
	for id := range ids {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			ds, err := c.getBundle(id)
			if err != nil {
				errChan <- err
				return
			}
			dsChan <- ds
		}(id)
	}
	wg.Wait()

	if len(errChan) > 0 {
		err := <-errChan
		return nil, nil, err
	}
	bundles := make([]*Bundle, 0, len(dsChan))
	for ds := range dsChan {
		bundles = append(bundles, ds)
	}

	return &BundleListResult{
		Bundles: bundles,
	}, nil, nil
}

// UpdateBundle creates or updates a bundle config. When updating, a Get should first be performed and the modified Bundle passed back.
func (c *ClusterConf) UpdateBundle(req *acomm.Request) (interface{}, *url.URL, error) {
	var args BundlePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Bundle == nil {
		return nil, nil, errors.New("missing arg: bundle")
	}
	args.Bundle.c = c

	if args.Bundle.ID == 0 {
		rand.Seed(time.Now().UnixNano())
		args.Bundle.ID = uint64(rand.Int63())
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
		return nil, nil, errors.New("missing arg: id")
	}
	if args.Serial == "" {
		return nil, nil, errors.New("missing arg: serial")
	}
	if args.IP == nil {
		return nil, nil, errors.New("missing arg: ip")
	}

	bundle, err := c.getBundle(args.ID)
	if err != nil {
		return nil, nil, err
	}

	if err := bundle.nodeHeartbeat(args.Serial, args.IP); err != nil {
		return nil, nil, err
	}

	return &BundlePayload{bundle}, nil, nil
}

func (c *ClusterConf) getBundle(id uint64) (*Bundle, error) {
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
	key := path.Join(bundlesPrefix, strconv.FormatUint(b.ID, 10))
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
	key := path.Join(bundlesPrefix, strconv.FormatUint(b.ID, 10))
	return b.c.kvDelete(key, b.ModIndex)
}

// update saves the core bundle config.
// It will not modify nodes.
func (b *Bundle) update() error {
	key := path.Join(bundlesPrefix, strconv.FormatUint(b.ID, 10), "config")

	_, err := b.c.kvUpdate(key, b.BundleConf, b.ModIndex)
	if err != nil {
		return err
	}

	// reload instead of just setting the new modIndex in case any nodes have also changed.
	return b.reload()
}

func (b *Bundle) nodeHeartbeat(serial string, ip net.IP) error {
	key := path.Join(bundlesPrefix, strconv.FormatUint(b.ID, 10), "nodes", serial)
	if err := b.c.kvEphemeral(key, ip, b.c.config.BundleTTL()); err != nil {
		return err
	}
	return b.reload()
}
