package main

import (
	"net"
	"time"

	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/pborman/uuid"
)

func (s *NodeHeartbeat) TestGetNodeInfo() {
	expectedIP, _, _ := net.ParseCIDR(s.metrics.Data.Network.Interfaces[0].Addrs[0].Addr)
	data, err := getNodeInfo(s.config, s.tracker, expectedIP)
	if !s.NoError(err) {
		return
	}
	s.Equal(expectedIP.String(), data.ID)
	s.Equal(s.metrics.Data.Memory.Virtual.Total, data.MemoryTotal)
	s.Equal(s.metrics.Data.Memory.Virtual.Available, data.MemoryFree)
	s.Equal(len(s.metrics.Data.CPU.Info), data.CPUCores)
	s.Equal(s.metrics.Data.CPU.Load, data.CPULoad)
	s.Equal(s.metrics.Data.Disk.Usage[0].Total, data.DiskTotal)
	s.Equal(s.metrics.Data.Disk.Usage[0].Free, data.DiskFree)
	s.WithinDuration(time.Now(), data.Heartbeat, time.Millisecond)
}

func (s *NodeHeartbeat) TestSendNodeHeartbeat() {
	data := &clusterconf.Node{
		ID:        uuid.New(),
		Heartbeat: time.Now(),
	}

	if !s.NoError(sendNodeHeartbeat(s.config, s.tracker, data)) {
		return
	}

	_, ok := s.clusterConf.Data.Nodes[data.ID]
	s.True(ok)
}
