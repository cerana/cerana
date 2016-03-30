package metrics

import "github.com/mistifyio/mistify/provider"

// Metrics is a provider of system info and metrics functionality.
type Metrics struct{}

// RegisterTasks registers all of Metric's task handlers with the server.
func (m *Metrics) RegisterTasks(server *provider.Server) {
	server.RegisterTask("metrics-cpu", m.CPU)
	server.RegisterTask("metrics-disk", m.Disk)
	server.RegisterTask("metrics-host", m.Host)
	server.RegisterTask("metrics-memory", m.Memory)
	server.RegisterTask("metrics-network", m.Network)
}
