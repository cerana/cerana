package namespace

import (
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

// Namespace is a provider of namespace functionality.
type Namespace struct {
	config  *provider.Config
	tracker *acomm.Tracker
}

// New creates a new instance of Namespace.
func New(config *provider.Config, tracker *acomm.Tracker) *Namespace {
	return &Namespace{
		config:  config,
		tracker: tracker,
	}
}

// RegisterTasks registers all of Namespaces's task handlers with the server.
func (n *Namespace) RegisterTasks(server *provider.Server) {
	server.RegisterTask("namespace-set-user", n.SetUser)
}
