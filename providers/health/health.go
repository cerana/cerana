package health

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

// Health is a provider of health checks.
type Health struct {
	config  *provider.Config
	tracker *acomm.Tracker
}

// New create a new instance of a Health provider.
func New(config *provider.Config, tracker *acomm.Tracker) *Health {
	return &Health{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of Health's task handlers with the server.
func (h *Health) RegisterTasks(server *provider.Server) {
	server.RegisterTask("health-uptime", h.Uptime)
	server.RegisterTask("health-file", h.File)
	server.RegisterTask("health-tcp-response", h.TCPResponse)
	server.RegisterTask("health-http-status", h.HTTPStatus)
}
