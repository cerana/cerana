package main

import (
	"time"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/pborman/uuid"
)

func (s *StatsPusher) TestGetNodeInfo() {
	data, err := s.statsPusher.getNodeInfo()
	if !s.NoError(err) {
		return
	}
	s.Equal(s.metrics.Data.Host.Hostname, data.ID)
	s.Equal(s.metrics.Data.Memory.Virtual.Total, data.MemoryTotal)
	s.Equal(s.metrics.Data.Memory.Virtual.Available, data.MemoryFree)
	s.Equal(len(s.metrics.Data.CPU.Info), data.CPUCores)
	s.Equal(s.metrics.Data.CPU.Load, data.CPULoad)
	s.Equal(s.metrics.Data.Disk.Usage[0].Total, data.DiskTotal)
	s.Equal(s.metrics.Data.Disk.Usage[0].Free, data.DiskFree)
	s.WithinDuration(time.Now(), data.Heartbeat, time.Millisecond)
}

func (s *StatsPusher) TestSendNodeHeartbeat() {
	data := &clusterconf.Node{
		ID:        uuid.New(),
		Heartbeat: time.Now(),
	}

	if !s.NoError(s.statsPusher.sendNodeHeartbeat(data)) {
		return
	}

	_, ok := s.clusterConf.Data.Nodes[data.ID]
	s.True(ok)
}
