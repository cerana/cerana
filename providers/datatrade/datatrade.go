package datatrade

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

// Provider is a provider of data import and export functionality.
type Provider struct {
	config  *Config
	tracker *acomm.Tracker
}

// New creates a new instance of Provider.
func New(config *Config, tracker *acomm.Tracker) *Provider {
	return &Provider{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of the provider task handlers with the server.
func (p *Provider) RegisterTasks(server *provider.Server) {
	server.RegisterTask("import-dataset", p.DatasetImport)
}
