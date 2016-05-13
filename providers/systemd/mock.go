package systemd

import (
	"errors"
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/coreos/go-systemd/dbus"
)

// MockSystemd is a mock version of the Systemd provider.
type MockSystemd struct {
	Data *MockSystemdData
}

// MockSystemdData is the in-memory data structure for the MockSystemd.
type MockSystemdData struct {
	Statuses  map[string]dbus.UnitStatus
	UnitFiles map[string]bool
}

// NewMockSystemd creates a new MockSystemd.
func NewMockSystemd() *MockSystemd {
	return &MockSystemd{
		Data: &MockSystemdData{
			Statuses:  make(map[string]dbus.UnitStatus),
			UnitFiles: make(map[string]bool),
		},
	}
}

// RegisterTasks registers the MockSystemd tasks.
func (s *MockSystemd) RegisterTasks(server *provider.Server) {
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

// Create creates a mock unit file.
func (s *MockSystemd) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := s.Data.UnitFiles[args.Name]; ok {
		return nil, nil, errors.New("unit file already exists")
	}
	s.Data.UnitFiles[args.Name] = true
	return nil, nil, nil
}

// Disable disables a mock service.
func (s *MockSystemd) Disable(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DisableArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	delete(s.Data.Statuses, args.Name)
	return nil, nil, nil
}

// Enable enables a mock service.
func (s *MockSystemd) Enable(req *acomm.Request) (interface{}, *url.URL, error) {
	var args EnableArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := s.Data.UnitFiles[args.Name]; !ok {
		return nil, nil, errors.New("No such file or directory")
	}

	s.Data.Statuses[args.Name] = dbus.UnitStatus{
		Name:      args.Name,
		LoadState: "Loaded",
	}
	return nil, nil, nil
}

// Get retrieves a mock service.
func (s *MockSystemd) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args GetArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	status, ok := s.Data.Statuses[args.Name]
	if !ok {
		return nil, nil, errors.New("No such file or directory")
	}
	return &GetResult{status}, nil, nil
}

// List lists mock services.
func (s *MockSystemd) List(req *acomm.Request) (interface{}, *url.URL, error) {
	list := make([]dbus.UnitStatus, 0, len(s.Data.Statuses))
	for _, status := range list {
		list = append(list, status)
	}
	return &ListResult{list}, nil, nil
}

// Remove removes a mock unit file.
func (s *MockSystemd) Remove(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RemoveArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := s.Data.Statuses[args.Name]; !ok {
		return nil, nil, errors.New("unit not found")
	}
	return nil, nil, nil
}

func (s *MockSystemd) action(req *acomm.Request) (interface{}, *url.URL, error) {
	var args ActionArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := s.Data.Statuses[args.Name]; !ok {
		return nil, nil, errors.New("unit not found")
	}
	return nil, nil, nil
}

// Restart restarts a mock service.
func (s *MockSystemd) Restart(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.action(req)
}

// Start starts a mock service.
func (s *MockSystemd) Start(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.action(req)
}

// Stop stops a mock service.
func (s *MockSystemd) Stop(req *acomm.Request) (interface{}, *url.URL, error) {
	return s.action(req)
}
