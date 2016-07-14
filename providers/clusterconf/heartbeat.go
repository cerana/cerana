package clusterconf

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"path"
	"strconv"

	"github.com/cerana/cerana/acomm"
)

const heartbeatPrefix = "heartbeats"

// DatasetHeartbeatArgs are arguments for updating a dataset node heartbeat.
type DatasetHeartbeatArgs struct {
	ID    string `json:"id"`
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}

// DatasetHeartbeat is dataset heartbeat information.
type DatasetHeartbeat struct {
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}

// DatasetHeartbeatList is the result of a ListDatasetHeartbeats.
type DatasetHeartbeatList struct {
	Heartbeats map[string]map[string]DatasetHeartbeat
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

	key := path.Join(heartbeatPrefix, datasetsPrefix, args.ID, args.IP.String())
	return nil, nil, c.kvEphemeral(key, args.InUse, c.config.DatasetTTL())
}

// ListDatasetHeartbeats returns a list of all active dataset heartbeats.
func (c *ClusterConf) ListDatasetHeartbeats(req *acomm.Request) (interface{}, *url.URL, error) {
	base := path.Join(heartbeatPrefix, datasetsPrefix)
	values, err := c.kvGetAll(base)
	if err != nil {
		return nil, nil, err
	}
	heartbeats := make(map[string]map[string]DatasetHeartbeat)
	for key, value := range values {
		if key == base {
			continue
		}
		// key: {base}/{id}/{ip}
		id := path.Base(path.Dir(key))
		ip := net.ParseIP(path.Base(key))
		var inUse bool
		if err := json.Unmarshal(value.Data, &inUse); err != nil {
			return nil, nil, err
		}
		if _, ok := heartbeats[id]; !ok {
			heartbeats[id] = make(map[string]DatasetHeartbeat)
		}
		heartbeats[id][ip.String()] = DatasetHeartbeat{IP: ip, InUse: inUse}
	}

	return DatasetHeartbeatList{heartbeats}, nil, nil
}

// BundleHeartbeatArgs are argumenst for updating a bundle heartbeat.
type BundleHeartbeatArgs struct {
	ID           uint64           `json:"id"`
	Serial       string           `json:"serial"`
	IP           net.IP           `json:"ip"`
	HealthErrors map[string]error `json:"healthErrors"`
}

// BundleHeartbeat is bundle heartbeat information.
type BundleHeartbeat struct {
	IP           net.IP           `json:"ip"`
	HealthErrors map[string]error `json:"healthErrors"`
}

// MarshalJSON marshals BundleHeartbeat into a JSON map, converting error
// values to strings.
func (b BundleHeartbeat) MarshalJSON() ([]byte, error) {
	type Alias BundleHeartbeat
	errors := make(map[string]string)
	for key, err := range b.HealthErrors {
		errors[key] = err.Error()
	}
	return json.Marshal(&struct {
		HealthErrors map[string]string `json:"healthErrors"`
		Alias
	}{
		HealthErrors: errors,
		Alias:        (Alias)(b),
	})
}

// UnmarshalJSON unmarshals JSON into a BundleHeartbeat, converting string
// values to errors.
func (b *BundleHeartbeat) UnmarshalJSON(data []byte) error {
	type Alias BundleHeartbeat
	aux := &struct {
		HealthErrors map[string]string `json:"healthErrors"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.HealthErrors = make(map[string]error)
	for key, errS := range aux.HealthErrors {
		b.HealthErrors[key] = errors.New(errS)
	}
	return nil
}

// BundleHeartbeats are a set of bundle heartbeats for a node.
type BundleHeartbeats map[string]BundleHeartbeat

// BundleHeartbeatList is the result of a ListBundleHeartbeats.
type BundleHeartbeatList struct {
	Heartbeats map[uint64]BundleHeartbeats `json:"heartbeats"`
}

// MarshalJSON marshals BundleHeartbeatList into a JSON map, converting uint
// keys to strings.
// TODO: Needed until go 1.7 is released
func (b BundleHeartbeatList) MarshalJSON() ([]byte, error) {
	type Alias BundleHeartbeatList
	hbs := make(map[string]BundleHeartbeats)
	for id, value := range b.Heartbeats {
		hbs[strconv.FormatUint(id, 10)] = value
	}
	return json.Marshal(&struct {
		Heartbeats map[string]BundleHeartbeats `json:"heartbeats"`
		Alias
	}{
		Heartbeats: hbs,
		Alias:      (Alias)(b),
	})
}

// UnmarshalJSON unmarshals JSON into a BundleHeartbeatList, converting string
// keys to uints.
// TODO: Needed until go 1.7 is released
func (b *BundleHeartbeatList) UnmarshalJSON(data []byte) error {
	type Alias BundleHeartbeatList
	aux := &struct {
		Heartbeats map[string]BundleHeartbeats `json:"heartbeats"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.Heartbeats = make(map[uint64]BundleHeartbeats)
	for idS, value := range aux.Heartbeats {
		id, err := strconv.ParseUint(idS, 10, 64)
		if err != nil {
			return err
		}
		b.Heartbeats[id] = value
	}
	return nil
}

// BundleHeartbeat registers a new node heartbeat that is using the dataset.
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

	heartbeat := BundleHeartbeat{
		IP:           args.IP,
		HealthErrors: args.HealthErrors,
	}

	key := path.Join(heartbeatPrefix, bundlesPrefix, strconv.FormatUint(args.ID, 10), args.Serial)
	return nil, nil, c.kvEphemeral(key, heartbeat, c.config.BundleTTL())
}

// ListBundleHeartbeats returns a list of all active bundle heartbeats.
func (c *ClusterConf) ListBundleHeartbeats(req *acomm.Request) (interface{}, *url.URL, error) {
	base := path.Join(heartbeatPrefix, bundlesPrefix)
	values, err := c.kvGetAll(base)
	if err != nil {
		return nil, nil, err
	}
	heartbeats := make(map[uint64]BundleHeartbeats)
	for key, value := range values {
		if key == base {
			continue
		}
		// key: {base}/{id}/{serial}
		id, err := strconv.ParseUint(path.Base(path.Dir(key)), 10, 64)
		if err != nil {
			return nil, nil, err
		}
		serial := path.Base(key)
		var hb BundleHeartbeat
		if err := json.Unmarshal(value.Data, &hb); err != nil {
			return nil, nil, err
		}
		if _, ok := heartbeats[id]; !ok {
			heartbeats[id] = make(BundleHeartbeats)
		}
		heartbeats[id][serial] = hb
	}

	return BundleHeartbeatList{heartbeats}, nil, nil
}
