package clusterconf

import "time"

// Node is current information about a hardware node.
type Node struct {
	ID          string `json:"id"`
	Heartbeat   time.Time
	MemoryTotal int64 `json:"memoryTotal"`
	MemoryFree  int64 `json:"memoryFree"`
	CPUTotal    int   `json:"cpuTotal"`
	CPUFree     int   `json:"cpuFree"`
	DiskTotal   int   `json:"diskTotal"`
	DiskFree    int   `json:"diskFree"`
}

// NodeHistory is a set of historical information for a node.
type NodeHistory map[time.Time]*Node

// NodesHistory is the historical information for multiple nodes.
type NodesHistory map[string]*NodeHistory
