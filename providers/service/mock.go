package service

import (
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
)

// Mock is a mock provider of service management functionality.
type Mock struct {
	Data MockData
}

// MockData is the in-memory data structure for the Mock.
type MockData struct {
	Services map[uint64]map[string]Service
}

// NewMock creates a new instance of Mock.
func NewMock() *Mock {
	return &Mock{
		Data: MockData{
			Services: make(map[uint64]map[string]Service),
		},
	}
}

// RegisterTasks registers all of the mock provider task handlers with the server.
func (m *Mock) RegisterTasks(server *provider.Server) {
	server.RegisterTask("service-create", m.Create)
	server.RegisterTask("service-get", m.Get)
	server.RegisterTask("service-list", m.List)
	server.RegisterTask("service-restart", m.Restart)
	server.RegisterTask("service-remove", m.Remove)
}

// Create creates a new mock service.
func (m *Mock) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.BundleID == 0 {
		return nil, nil, errors.New("missing arg: bundleID")
	}
	if len(args.Exec) == 0 {
		return nil, nil, errors.New("missing arg: exec")
	}
	if args.Dataset == "" {
		return nil, nil, errors.New("missing arg: dataset")
	}

	if _, ok := m.Data.Services[args.BundleID]; !ok {
		m.Data.Services[args.BundleID] = make(map[string]Service)
	}
	m.Data.Services[args.BundleID][args.ID] = Service{
		ID:          args.ID,
		BundleID:    args.BundleID,
		Description: args.Description,
		Uptime:      time.Minute,
		ActiveState: "Running",
		Exec:        args.Exec,
		UID:         uint64(rand.Int63n(60000)),
		GID:         uint64(rand.Int63n(60000)),
		Env:         args.Env,
	}
	return GetResult{m.Data.Services[args.BundleID][args.ID]}, nil, nil
}

// Get retrieves a mock service.
func (m *Mock) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.BundleID == 0 {
		return nil, nil, errors.New("missing arg: bundleID")
	}

	bundle, ok := m.Data.Services[args.BundleID]
	if !ok {
		return nil, nil, errors.New("could not find service")
	}
	service, ok := bundle[args.ID]
	if !ok {
		return nil, nil, errors.New("could not find service")
	}

	return GetResult{service}, nil, nil
}

// List lists all mock services.
func (m *Mock) List(req *acomm.Request) (interface{}, *url.URL, error) {
	services := []Service{}
	for _, bundle := range m.Data.Services {
		for _, service := range bundle {
			services = append(services, service)
		}
	}

	return ListResult{services}, nil, nil
}

// Restart restarts a mock service.
func (m *Mock) Restart(req *acomm.Request) (interface{}, *url.URL, error) {
	return nil, nil, nil
}

// Remove removes a mock service.
func (m *Mock) Remove(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RemoveArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.ID == "" {
		return nil, nil, errors.New("missing arg: id")
	}
	if args.BundleID == 0 {
		return nil, nil, errors.New("missing arg: bundleID")
	}

	if bundle, ok := m.Data.Services[args.BundleID]; ok {
		delete(bundle, args.ID)
	}

	return nil, nil, nil
}

// ClearData clears all mock data.
func (m *Mock) ClearData() {
	m.Data.Services = make(map[uint64]map[string]Service)
}

// Add is a convenience method to directly add a mock service.
func (m *Mock) Add(service Service) {
	if _, ok := m.Data.Services[service.BundleID]; !ok {
		m.Data.Services[service.BundleID] = make(map[string]Service)
	}
	m.Data.Services[service.BundleID][service.ID] = service
}
