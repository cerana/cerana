package clusterconf

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/pborman/uuid"
)

const datasetsPrefix string = "datasets"

// Dataset is information about a dataset.
type Dataset struct {
	c                 *ClusterConf
	ID                string `json:"id"`
	Parent            string `json:"parent"`
	ParentSameMachine bool   `json:"parentSameMachine"`
	ReadOnly          bool   `json:"readOnly"`
	NFS               bool   `json:"nfs"`
	Redundancy        uint64 `json:"redundancy"`
	Quota             uint64 `json:"quota"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}

// DatasetPayload can be used for task args or result when a dataset object
// needs to be sent.
type DatasetPayload struct {
	Dataset *Dataset `json:"dataset"`
}

// DatasetListResult is the result for listing datasets.
type DatasetListResult struct {
	Datasets []*Dataset `json:"datasets"`
}

// GetDataset retrieves a dataset.
func (c *ClusterConf) GetDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.Newv("missing arg: id", map[string]interface{}{"args": args})
	}

	dataset, err := c.getDataset(args.ID)
	if err != nil {
		return nil, nil, err
	}
	return &DatasetPayload{dataset}, nil, nil
}

// ListDatasets returns a list of all Datasets.
func (c *ClusterConf) ListDatasets(req *acomm.Request) (interface{}, *url.URL, error) {
	keys, err := c.kvKeys(datasetsPrefix)
	if err != nil {
		return nil, nil, err
	}
	// extract and deduplicate the dataset ids
	ids := make(map[string]bool)
	keyFormat := filepath.Join(datasetsPrefix, "%s")
	for _, key := range keys {
		var id string
		_, err := fmt.Sscanf(key, keyFormat, &id)
		if err != nil {
			return nil, nil, errors.Newv("failed to extract valid dataset id", map[string]interface{}{"key": key, "keyFormat": keyFormat})
		}
		ids[id] = true
	}

	var wg sync.WaitGroup
	dsChan := make(chan *Dataset, len(ids))
	errChan := make(chan error, len(ids))
	for id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			ds, err := c.getDataset(id)
			if err != nil {
				errChan <- err
				return
			}
			dsChan <- ds
		}(id)
	}
	wg.Wait()

	close(dsChan)
	close(errChan)

	if len(errChan) > 0 {
		err := <-errChan
		return nil, nil, err
	}
	datasets := make([]*Dataset, 0, len(dsChan))
	for ds := range dsChan {
		datasets = append(datasets, ds)
	}

	return &DatasetListResult{
		Datasets: datasets,
	}, nil, nil
}

// UpdateDataset creates or updates a dataset config. When updating, a Get should first be performed and the modified Dataset passed back.
func (c *ClusterConf) UpdateDataset(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DatasetPayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Dataset == nil {
		return nil, nil, errors.Newv("missing arg: dataset", map[string]interface{}{"args": args})
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
		return nil, nil, errors.Newv("missing arg: id", map[string]interface{}{"args": args})
	}

	dataset, err := c.getDataset(args.ID)
	if err != nil {
		if err.Error() == "dataset config not found" {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	return nil, nil, dataset.delete()
}

func (c *ClusterConf) getDataset(id string) (*Dataset, error) {
	dataset := &Dataset{
		c:  c,
		ID: id,
	}
	if err := dataset.reload(); err != nil {
		return nil, err
	}
	return dataset, nil
}

func (d *Dataset) reload() error {
	var err error
	key := path.Join(datasetsPrefix, d.ID, "config")
	value, err := d.c.kvGet(key)
	if err != nil {
		if strings.Contains(err.Error(), "key not found") {
			err = errors.Newv("dataset config not found", map[string]interface{}{"datasetID": d.ID})
		}
		return err
	}

	if err = json.Unmarshal(value.Data, &d); err != nil {
		return errors.Wrapv(err, map[string]interface{}{"json": string(value.Data)})
	}
	d.ModIndex = value.Index
	return nil
}

func (d *Dataset) delete() error {
	key := path.Join(datasetsPrefix, d.ID)
	return errors.Wrapv(d.c.kvDelete(key, d.ModIndex), map[string]interface{}{"datasetID": d.ID})
}

// update saves the core dataset config.
func (d *Dataset) update() error {
	key := path.Join(datasetsPrefix, d.ID, "config")

	index, err := d.c.kvUpdate(key, d, d.ModIndex)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"datasetID": d.ID})
	}
	d.ModIndex = index
	return nil
}
