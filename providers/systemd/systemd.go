package systemd

import (
	"github.com/mistifyio/mistify/acomm"
	"github.com/mistifyio/mistify/provider"
)

// Unit start modes.
const (
	ModeReplace    = "replace"
	ModeFail       = "fail"
	ModeIsolate    = "isolate"
	ModeIgnoreDeps = "ignore-dependencies"
	ModeIgnoreReqs = "ignore-requirements"
)

// Systemd is a provider of systemd functionality.
type Systemd struct {
	config  *Config
	tracker *acomm.Tracker
}

// New creates a new instance of Systemd.
func New(config *provider.Config) *Systemd {
	return &Systemd{
		config: &Config{config},
	}
}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (s *Systemd) RegisterTasks(server *provider.Server) {
	server.RegisterTask("systemd-disable", s.Disable)
	server.RegisterTask("systemd-enable", s.Enable)
	server.RegisterTask("systemd-get", s.Get)
	server.RegisterTask("systemd-list", s.List)
	server.RegisterTask("systemd-restart", s.Restart)
	server.RegisterTask("systemd-start", s.Start)
	server.RegisterTask("systemd-stop", s.Stop)
}
