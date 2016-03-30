package systemd

import (
	"github.com/coreos/go-systemd/dbus"
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
	config *Config
	dconn  *dbus.Conn
}

// New creates a new instance of Systemd.
func New(config *Config) (*Systemd, error) {
	dconn, err := dbus.New()
	if err != nil {
		return nil, err
	}
	return &Systemd{
		config: config,
		dconn:  dconn,
	}, nil
}

// RegisterTasks registers all of Systemd's task handlers with the server.
func (s *Systemd) RegisterTasks(server *provider.Server) {
	server.RegisterTask("systemd-create", s.Create)
	server.RegisterTask("systemd-disable", s.Disable)
	server.RegisterTask("systemd-enable", s.Enable)
	server.RegisterTask("systemd-get", s.Get)
	server.RegisterTask("systemd-list", s.List)
	server.RegisterTask("systemd-remove", s.Remove)
	server.RegisterTask("systemd-restart", s.Restart)
	server.RegisterTask("systemd-start", s.Start)
	server.RegisterTask("systemd-stop", s.Stop)
}
