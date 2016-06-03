package clusterconf

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"time"

	"github.com/cerana/cerana/acomm"
)

const (
	nodesPrefix      string = "nodes"
	historicalPrefix string = "historical"
)

// Node is current information about a hardware node.
type Node struct {
	c           *ClusterConf
	ID          string    `json:"id"`
	Heartbeat   time.Time `json:"heartbeat"`
	MemoryTotal int64     `json:"memoryTotal"`
	MemoryFree  int64     `json:"memoryFree"`
	CPUTotal    int       `json:"cpuTotal"`
	CPUFree     int       `json:"cpuFree"`
	DiskTotal   int       `json:"diskTotal"`
	DiskFree    int       `json:"diskFree"`
}

// NodeHistory is a set of historical information for a node.
type NodeHistory map[time.Time]Node

// NodesHistory is the historical information for multiple nodes.
type NodesHistory map[string]NodeHistory

// NodeHistoryArgs are arguments for filtering the historical results for nodes.
type NodeHistoryArgs struct {
	IDs    []string  `json:"ids"`
	Before time.Time `json:"before"`
	After  time.Time `json:"after"`
}

// NodePayload can be used for task args or result when a node object needs to
// be sent.
type NodePayload struct {
	Node *Node `json:"node"`
}

// NodesHistoryResult is the result from the GetNodesHistory handler.
type NodesHistoryResult struct {
	History *NodesHistory `json:"history"`
}

type nodeFilter func(Node) bool

// NodeHeartbeat records a new node heartbeat.
func (c *ClusterConf) NodeHeartbeat(req *acomm.Request) (interface{}, *url.URL, error) {
	var args NodePayload
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Node == nil {
		return nil, nil, errors.New("missing arg: node")
	}
	args.Node.c = c

	return nil, nil, args.Node.update()
}

// GetNode returns the latest information about a node.
func (c *ClusterConf) GetNode(req *acomm.Request) (interface{}, *url.URL, error) {
	var args IDArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}

	node, err := c.getNode(args.ID)
	return &NodePayload{node}, nil, err
}

// GetNodesHistory gets the heartbeat history for one or more nodes.
func (c *ClusterConf) GetNodesHistory(req *acomm.Request) (interface{}, *url.URL, error) {
	var args NodeHistoryArgs
	if err := req.UnmarshalArgs(args); err != nil {
		return nil, nil, err
	}

	history, err := c.getNodesHistory(
		nodeFilterID(args.IDs...),
		nodeFilterHeartbeat(args.Before, args.After),
	)
	if err != nil {
		return nil, nil, err
	}
	return &NodesHistoryResult{history}, nil, nil

}

func (c *ClusterConf) getNode(id string) (*Node, error) {
	node := &Node{}
	key := path.Join(nodesPrefix, id)
	value, err := c.kvGet(key)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(value.Data, node); err != nil {
		return nil, err
	}
	node.c = c
	return node, nil
}

func nodeFilterID(ids ...string) nodeFilter {
	return func(n Node) bool {
		for _, id := range ids {
			if n.ID == id {
				return true
			}
		}
		return false
	}
}

func nodeFilterHeartbeat(before, after time.Time) nodeFilter {
	return func(n Node) bool {
		if !before.IsZero() && !n.Heartbeat.Before(before) {
			return false
		}
		if !after.IsZero() && !n.Heartbeat.After(after) {
			return false
		}
		return true
	}
}

func (c *ClusterConf) getNodesHistory(filters ...nodeFilter) (*NodesHistory, error) {
	values, err := c.kvGetAll(historicalPrefix)
	if err != nil {
		return nil, err
	}

	history := make(NodesHistory)
	for _, value := range values {
		var node Node
		if err := json.Unmarshal(value.Data, &node); err != nil {
			return nil, err
		}

		for _, fn := range filters {
			if !fn(node) {
				continue
			}
		}

		nodeHistory, ok := history[node.ID]
		if !ok {
			nodeHistory = make(NodeHistory)
			history[node.ID] = nodeHistory
		}
		nodeHistory[node.Heartbeat] = node
	}

	return &history, nil
}

func (n *Node) update() error {
	currentKey := path.Join(nodesPrefix, n.ID)
	historicalKey := path.Join(historicalPrefix, n.ID, n.Heartbeat.Format(time.RFC3339))

	multiRequest := acomm.NewMultiRequest(n.c.tracker, 0)

	currentReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-ephemeral",
		Args: map[string]interface{}{
			"key":   currentKey,
			"value": n,
			"ttl":   n.c.config.NodeTTL(),
		},
	})
	if err != nil {
		return err
	}
	historicalReq, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "kv-update",
		Args: map[string]interface{}{
			"key":   historicalKey,
			"value": n,
		},
	})
	if err != nil {
		return err
	}

	requests := map[string]*acomm.Request{
		"current":    currentReq,
		"historical": historicalReq,
	}

	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			continue
		}
		if err := acomm.Send(n.c.config.CoordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			continue
		}
	}

	responses := multiRequest.Responses()
	for name := range requests {
		resp, ok := responses[name]
		if !ok {
			return fmt.Errorf("failed to send request: %s", name)
		}
		if resp.Error != nil {
			return fmt.Errorf("request failed: %s: %s", name, resp.Error)
		}
	}

	return nil
}
