package systemd

import "github.com/mistifyio/mistify/provider"

// Systemd is a provider of systemd functionality
type Systemd struct{}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (s *Systemd) RegisterTasks(server *provider.Server) {
	server.RegisterTask("systemd-list-units", s.ListUnits)
	server.RegisterTask("systemd-get", s.Get)
}
