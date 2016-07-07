package clusterconf

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"path"

	"github.com/cerana/cerana/acomm"
)

const heartbeatPrefix = "heartbeats"

// DatasetHeartbeatArgs are arguments for updating a dataset node heartbeat.
type DatasetHeartbeatArgs struct {
	ID    string `json:"id"`
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}

type DatasetHeartbeat struct {
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}

type DatasetHeartbeatList struct {
	Heartbeats map[string][]DatasetHeartbeat
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

func (c *ClusterConf) ListDatasetHeartbeats(req *acomm.Request) (interface{}, *url.URL, error) {
	base := path.Join(heartbeatPrefix, datasetsPrefix)
	values, err := c.kvGetAll(base)
	if err != nil {
		return nil, nil, err
	}
	heartbeats := make(map[string][]DatasetHeartbeat)
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
			heartbeats[id] = []DatasetHeartbeat{}
		}
		heartbeats[id] = append(heartbeats[id], DatasetHeartbeat{IP: ip, InUse: inUse})
	}

	return DatasetHeartbeatList{heartbeats}, nil, nil
}
