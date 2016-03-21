package systemd

import "github.com/mistifyio/mistify/provider"

// Unit start modes
const (
	ModeReplace    = "replace"
	ModeFail       = "fail"
	ModeIsolate    = "isolate"
	ModeIgnoreDeps = "ignore-dependencies"
	ModeIgnoreReqs = "ignore-requirements"
)

// Systemd is a provider of systemd functionality
type Systemd struct{}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (s *Systemd) RegisterTasks(server *provider.Server) {
	server.RegisterTask("systemd-get", s.Get)
	server.RegisterTask("systemd-list", s.List)
	server.RegisterTask("systemd-restart", s.Start)
	server.RegisterTask("systemd-start", s.Start)
	server.RegisterTask("systemd-stop", s.Start)
}
