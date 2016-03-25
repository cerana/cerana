package metrics

import "github.com/mistifyio/mistify/provider"

type Metrics struct{}

func (m *Metrics) RegisterTasks(server *provider.Server) {
	server.RegisterTask("metrics-cpu", m.CPU)
	server.RegisterTask("metrics-disk", m.Disk)
	server.RegisterTask("metrics-host", m.Host)
	server.RegisterTask("metrics-memory", m.Memory)
	server.RegisterTask("metrics-network", m.Network)
}
