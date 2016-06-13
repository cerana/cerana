package clusterconf

import (
	"encoding/json"
	"errors"
	"fmt"
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

// BundleDatasetType is the type of dataset to be used in a bundle.
type BundleDatasetType int

// Valid bundle dataset types
const (
	RWZFS = iota
	TempZFS
	RAMDisk
)

// Bundle is information about a bundle of services.
type Bundle struct {
	BundleConf
	c *ClusterConf
	// Nodes contains the set of nodes on which the dataset is currently in use.
	// The map keys are serials.
	Nodes map[string]net.IP `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}

// BundleConf is the configuration of a bundle.
type BundleConf struct {
	ID         uint64                   `json:"id"`
	Datasets   map[string]BundleDataset `json:"datasets"`
	Services   map[string]BundleService `json:"services"`
	Redundancy int                      `json:"redundancy"`
	Ports      BundlePorts              `json:"ports"`
}

// BundlePorts is a map of port numbers to port information.
type BundlePorts map[int]BundlePort

// MarshalJSON marshals BundlePorts into a JSON map, converting int keys to
// strings.
func (p BundlePorts) MarshalJSON() ([]byte, error) {
	ports := make(map[string]BundlePort)
	for port, value := range p {
		ports[strconv.Itoa(port)] = value
	}
	return json.Marshal(ports)
}

// UnmarshalJSON unmarshals JSON into a BundlePorts, converting string keys to
// ints.
func (p BundlePorts) UnmarshalJSON(data []byte) error {
	ports := make(map[string]BundlePort)
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
	Name  string            `json:"name"`
	ID    string            `json:"id"`
	Type  BundleDatasetType `json:"type"`
	Quota int               `json:"type"`
}

func (d BundleDataset) overlayOn(base *Dataset) (BundleDataset, error) {
	if d.ID != base.ID {
		return d, errors.New("dataset ids do not match")
	}

	// overlay data
	if d.Quota <= 0 {
		d.Quota = base.Quota
	}

	return d, nil
}

// BundleService is configuration overrides for a service of a bundle and
// associated bundles.
type BundleService struct {
	ServiceConf
	Datasets map[string]ServiceDataset `json:"datasets"`
}

func (s BundleService) overlayOn(base *Service) (BundleService, error) {
	if s.ID != base.ID {
		return s, errors.New("service ids do not match")
	}

	// maps are pointers, so need to be duplicated separately.
	result := s
	result.Datasets = make(map[string]ServiceDataset)
	for k, v := range s.Datasets {
		result.Datasets[k] = v
	}
	result.HealthChecks = make(map[string]HealthCheck)
	for k, v := range s.HealthChecks {
		result.HealthChecks[k] = v
	}
	result.Env = make(map[string]string)
	for k, v := range s.Env {
		result.Env[k] = v
	}

	// overlay data
	if result.Dataset == "" {
		result.Dataset = base.Dataset
	}

	if result.Limits.CPU <= 0 {
		result.Limits.CPU = base.Limits.CPU
	}
	if result.Limits.Memory <= 0 {
		result.Limits.Memory = base.Limits.Memory
	}
	if result.Limits.Processes <= 0 {
		result.Limits.Processes = base.Limits.Processes
	}

	for id, hc := range base.HealthChecks {
		_, ok := result.HealthChecks[id]
		if !ok {
			result.HealthChecks[id] = hc
			continue
		}
	}
	for key, val := range base.Env {
		_, ok := result.Env[key]
		if !ok {
			result.Env[key] = val
		}
	}

	return result, nil
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

// DeleteBundleArgs are args for bundle delete task.
type DeleteBundleArgs struct {
	ID uint64 `json:"id"`
}

// GetBundleArgs are args for retrieving a bundle.
type GetBundleArgs struct {
	ID              uint64 `json:"id"`
	CombinedOverlay bool   `json:"overlay"`
}

// ListBundleArgs are args for retrieving a bundle list.
type ListBundleArgs struct {
	CombinedOverlay bool `json:"overlay"`
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
	var args GetBundleArgs
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
	if args.CombinedOverlay {
		bundle, err = bundle.combinedOverlay()
		if err != nil {
			return nil, nil, err
		}
	}
	return &BundlePayload{bundle}, nil, nil
}

// ListBundles retrieves a list of all bundles.
func (c *ClusterConf) ListBundles(req *acomm.Request) (interface{}, *url.URL, error) {
	var args ListBundleArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	keys, err := c.kvKeys(bundlesPrefix)
	if err != nil {
		return nil, nil, err
	}
	// extract and deduplicate the bundle ids
	ids := make(map[uint64]bool)
	for _, key := range keys {
		// keys are full paths and include all child keys.
		// e.g. {prefix}/{id}/{rest/of/path}
		idS := strings.Split(strings.TrimPrefix(key, bundlesPrefix), "/")[0]
		id, err := strconv.ParseUint(idS, 10, 64)
		if err != nil {
			return nil, nil, errors.New("invalid bundle id")
		}
		ids[id] = true
	}

	var wg sync.WaitGroup
	bundleChan := make(chan *Bundle, len(ids))
	defer close(bundleChan)
	errChan := make(chan error, len(ids))
	defer close(errChan)
	for id := range ids {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			bundle, err := c.getBundle(id)
			if err != nil {
				errChan <- err
				return
			}
			if args.CombinedOverlay {
				bundle, err = bundle.combinedOverlay()
				if err != nil {
					errChan <- err
					return
				}
			}
			bundleChan <- bundle
		}(id)
	}
	wg.Wait()

	if len(errChan) > 0 {
		err := <-errChan
		return nil, nil, err
	}
	bundles := make([]*Bundle, 0, len(bundleChan))
	for bundle := range bundleChan {
		bundles = append(bundles, bundle)
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
	var args DeleteBundleArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == 0 {
		return nil, nil, errors.New("missing arg: id")
	}

	bundle, err := c.getBundle(args.ID)
	if err != nil {
		if err.Error() == "bundle config not found" {
			return nil, nil, nil
		}
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
		BundleConf: BundleConf{ID: id},
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
	if err = json.Unmarshal(config.Data, &b.BundleConf); err != nil {
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

// combinedOverlay will create a new *Bundle object containing the base configurations of datasets and services with the bundle values overlayed on top.
// Note: Attempting to save a combined overlay bundle will result in an error.
func (b *Bundle) combinedOverlay() (*Bundle, error) {
	var wg sync.WaitGroup
	errorChan := make(chan error, len(b.Datasets)+len(b.Services))
	defer close(errorChan)

	// duplicate bundle
	result := &Bundle{
		BundleConf: b.BundleConf,
		Nodes:      make(map[string]net.IP),
	}
	for k, v := range b.Nodes {
		result.Nodes[k] = v
	}
	result.Datasets = make(map[string]BundleDataset)
	for k, v := range b.Datasets {
		result.Datasets[k] = v
	}
	result.Services = make(map[string]BundleService)
	for k, v := range b.Services {
		result.Services[k] = v
	}
	result.Ports = make(BundlePorts)
	for k, v := range b.Ports {
		result.Ports[k] = v
	}

	for i, d := range b.Datasets {
		wg.Add(1)
		go func(id string, bd BundleDataset) {
			defer wg.Done()
			dataset, err := b.c.getDataset(id)
			if err != nil {
				errorChan <- err
				return
			}
			combined, err := bd.overlayOn(dataset)
			if err != nil {
				errorChan <- err
				return
			}
			result.Datasets[id] = combined
		}(i, d)
	}

	for i, s := range b.Services {
		wg.Add(1)
		go func(id string, bs BundleService) {
			defer wg.Done()
			service, err := b.c.getService(id)
			if err != nil {
				errorChan <- err
				return
			}
			combined, err := bs.overlayOn(service)
			if err != nil {
				errorChan <- err
				return
			}
			result.Services[id] = combined
		}(i, s)
	}

	wg.Wait()

	if len(errorChan) == 0 {
		return result, nil
	}

	errors := make([]error, len(errorChan))
Loop:
	for {
		select {
		case err := <-errorChan:
			errors = append(errors, err)
		default:
			break Loop
		}
	}
	return nil, fmt.Errorf("bundle overlay failed: %+v", errors)
}

func (b *Bundle) nodeHeartbeat(serial string, ip net.IP) error {
	key := path.Join(bundlesPrefix, strconv.FormatUint(b.ID, 10), "nodes", serial)
	if err := b.c.kvEphemeral(key, ip, b.c.config.BundleTTL()); err != nil {
		return err
	}
	return b.reload()
}
