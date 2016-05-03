package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
)

func (s *statsPusher) nodeHeartbeat() error {
	node, err := s.getNodeInfo()
	if err != nil {
		return err
	}
	return s.sendNodeHeartbeat(node)
}

func (s *statsPusher) getNodeInfo() (*clusterconf.Node, error) {
	var err error

	tasks := map[string]interface{}{
		"metrics-cpu":    &metrics.CPUResult{},
		"metrics-disk":   &metrics.DiskResult{},
		"metrics-host":   &host.InfoStat{},
		"metrics-memory": &metrics.MemoryResult{},
	}
	requests := make(map[string]*acomm.Request)
	for task := range tasks {
		requests[task], err = acomm.NewRequest(acomm.RequestOptions{Task: "task"})
		if err != nil {
			return nil, err
		}
	}

	multiRequest := acomm.NewMultiRequest(s.tracker, s.config.requestTimeout())
	for name, req := range requests {
		if err := multiRequest.AddRequest(name, req); err != nil {
			break
		}
		if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
			multiRequest.RemoveRequest(req)
			break
		}
	}

	responses := multiRequest.Responses()
	for name := range requests {
		resp, ok := responses[name]
		if !ok {
			return nil, fmt.Errorf("failed to send request: %s", name)
		}
		if resp.Error != nil {
			return nil, fmt.Errorf("request failed: %s: %s", name, resp.Error)
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
		ID:          tasks["metrics-host"].(*host.InfoStat).Hostname,
		Heartbeat:   time.Now(),
		MemoryTotal: tasks["metrics-memory"].(*metrics.MemoryResult).Virtual.Total,
		MemoryFree:  tasks["metrics-memory"].(*metrics.MemoryResult).Virtual.Available,
		CPUCores:    len(tasks["metrics-cpu"].(*metrics.CPUResult).Info),
		CPULoad:     tasks["metrics-cpu"].(*metrics.CPUResult).Load,
		DiskTotal:   usage.Total,
		DiskFree:    usage.Free,
	}, nil
}

func (s *statsPusher) sendNodeHeartbeat(data *clusterconf.Node) error {
	doneChan := make(chan error, 1)
	rh := func(_ *acomm.Request, resp *acomm.Response) {
		doneChan <- resp.Error
	}
	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task:           "node-heartbeat",
		TaskURL:        s.config.heartbeatURL(),
		ResponseHook:   s.tracker.URL(),
		Args:           data,
		SuccessHandler: rh,
		ErrorHandler:   rh,
	})
	if err != nil {
		return err
	}

	if err := s.tracker.TrackRequest(req, s.config.requestTimeout()); err != nil {
		return err
	}
	if err := acomm.Send(s.config.coordinatorURL(), req); err != nil {
		_ = s.tracker.RemoveRequest(req)
		return err
	}

	return <-doneChan
}
