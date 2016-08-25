package main

import (
	"net"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/tick"
	"github.com/shirou/gopsutil/disk"
)

func nodeHeartbeat(config tick.Configer, tracker *acomm.Tracker) error {
	ip, err := tick.GetIP(config, tracker)
	if err != nil {
		return err
	}

	node, err := getNodeInfo(config, tracker, ip)
	if err != nil {
		return err
	}
	return sendNodeHeartbeat(config, tracker, node)
}

func getNodeInfo(config tick.Configer, tracker *acomm.Tracker, ip net.IP) (*clusterconf.Node, error) {
	var err error
	tasks := map[string]interface{}{
		"metrics-cpu":    &metrics.CPUResult{},
		"metrics-disk":   &metrics.DiskResult{},
		"metrics-memory": &metrics.MemoryResult{},
	}
	requests := make(map[string]*acomm.Request)
	for task := range tasks {
		requests[task], err = acomm.NewRequest(acomm.RequestOptions{Task: task})
		if err != nil {
			return nil, err
		}
	}

	multiRequest := acomm.NewMultiRequest(tracker, config.RequestTimeout())
	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			return nil, err
		}
		if err := acomm.Send(config.NodeDataURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			return nil, err
		}
	}

	responses := multiRequest.Responses()
	for name := range requests {
		resp := responses[name]
		if resp.Error != nil {
			return nil, errors.Wrapv(errors.ResetStack(resp.Error), map[string]interface{}{"task": name})
		}
		if err := resp.UnmarshalResult(tasks[name]); err != nil {
			return nil, err
		}
	}

	usages := tasks["metrics-disk"].(*metrics.DiskResult).Usage
	var usage *disk.UsageStat
	for _, u := range usages {
		if u.Path == "/" {
			usage = u
			break
		}
	}
	if usage == nil {
		return nil, errors.New("failed to determine disk usage")
	}

	return &clusterconf.Node{
		ID:          ip.String(),
		Heartbeat:   time.Now(),
		MemoryTotal: tasks["metrics-memory"].(*metrics.MemoryResult).Virtual.Total,
		MemoryFree:  tasks["metrics-memory"].(*metrics.MemoryResult).Virtual.Available,
		CPUCores:    len(tasks["metrics-cpu"].(*metrics.CPUResult).Info),
		CPULoad:     tasks["metrics-cpu"].(*metrics.CPUResult).Load,
		DiskTotal:   usage.Total,
		DiskFree:    usage.Free,
	}, nil
}

func sendNodeHeartbeat(config tick.Configer, tracker *acomm.Tracker, data *clusterconf.Node) error {
	opts := acomm.RequestOptions{
		Task: "node-heartbeat",
		Args: &clusterconf.NodePayload{Node: data},
	}
	resp, err := tracker.SyncRequest(config.ClusterDataURL(), opts, config.RequestTimeout())
	if err != nil {
		return err
	}
	return errors.ResetStack(resp.Error)
}
