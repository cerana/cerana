package namespace

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/provider"
)

// Mock is a mock Namespace provider.
type Mock struct {
	Data MockData
}

// MockData is the in-memory data structure for a Mock.
type MockData struct {
	SetUserErr error
}

// NewMock creates a new instance of Mock.
func NewMock() *Mock {
	return &Mock{}
}

// RegisterTasks registers all of Mock's task handlers.
func (n *Mock) RegisterTasks(server *provider.Server) {
	server.RegisterTask("namespace-set-user", n.SetUser)
}

// SetUser sets mock uid and gid mappings.
func (n *Mock) SetUser(req *acomm.Request) (interface{}, *url.URL, error) {
	return nil, nil, errors.Wrap(n.Data.SetUserErr)
}
