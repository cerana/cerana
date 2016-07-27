package clusterconf

import (
	"encoding/json"
	"errors"
	"net/url"
	"path"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/shirou/gopsutil/load"
)

const (
	nodesPrefix      string = "nodes"
	historicalPrefix string = "historical"
)

// Node is current information about a hardware node.
type Node struct {
	c           *ClusterConf
	ID          string       `json:"id"`
	Heartbeat   time.Time    `json:"heartbeat"`
	MemoryTotal uint64       `json:"memoryTotal"`
	MemoryFree  uint64       `json:"memoryFree"`
	CPUCores    int          `json:"cpuCores"`
	CPULoad     load.AvgStat `json:"cpuLoad"`
	DiskTotal   uint64       `json:"diskTotal"`
	DiskFree    uint64       `json:"diskFree"`
}

// NodeHistory is a set of historical information for a node.
type NodeHistory map[time.Time]*Node

// NodesHistory is the historical information for multiple nodes.
type NodesHistory map[string]NodeHistory

// NodeHistoryArgs are arguments for filtering the historical results for nodes.
type NodeHistoryArgs struct {
	IDs    []string  `json:"ids"`
	Before time.Time `json:"before"`
	After  time.Time `json:"after"`
}

// ListNodesResult is the result of ListNodes.
type ListNodesResult struct {
	Nodes []Node `json:"nodes"`
}

// NodePayload can be used for task args or result when a node object needs to
// be sent.
type NodePayload struct {
	Node *Node `json:"node"`
}

// NodesHistoryResult is the result from the GetNodesHistory handler.
type NodesHistoryResult struct {
	History NodesHistory `json:"history"`
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

// ListNodes list all current nodes.
func (c *ClusterConf) ListNodes(req *acomm.Request) (interface{}, *url.URL, error) {
	nodes, err := c.getNodes()
	if err != nil {
		return nil, nil, err
	}
	return &ListNodesResult{nodes}, nil, nil
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
	return &NodesHistoryResult{*history}, nil, nil

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

func (c *ClusterConf) getNodes() ([]Node, error) {
	values, err := c.kvGetAll(nodesPrefix)
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(values))
	for key, value := range values {
		if key == nodesPrefix {
			continue
		}
		var node Node
		if err := json.Unmarshal(value.Data, node); err != nil {
			return nil, err
		}
		node.c = c
		nodes = append(nodes, node)
	}

	return nodes, nil
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
		nodeHistory[node.Heartbeat] = &node
	}

	return &history, nil
}

func (n *Node) update() error {
	currentKey := path.Join(nodesPrefix, n.ID)
	historicalKey := path.Join(historicalPrefix, n.ID, n.Heartbeat.Format(time.RFC3339))

	if err := n.c.kvEphemeral(currentKey, n, n.c.config.NodeTTL()); err != nil {
		return err
	}

	if _, err := n.c.kvUpdate(historicalKey, n, 0); err != nil {
		return err
	}

	return nil
}
