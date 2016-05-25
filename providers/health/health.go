package health

import "github.com/cerana/cerana/provider"

// Health is a provider of health checks.
type Health struct{}

// RegisterTasks registers all of Health's task handlers with the server.
func (h *Health) RegisterTasks(server *provider.Server) {
	/*
		server.RegisterTask("health-uptime", h.Uptime)
		server.Registertask("health-file-exists", h.FileExists)
		server.RegisterTask("health-resp-contains", h.RespContains)
	*/
	server.RegisterTask("health-http-status", h.HTTPStatus)
}
