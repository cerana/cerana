package systemd

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/coreos/go-systemd/dbus"
	"github.com/coreos/go-systemd/unit"
)

// MockSystemd is a mock version of the Systemd provider.
type MockSystemd struct {
	Data *MockSystemdData
}

// MockSystemdData is the in-memory data structure for the MockSystemd.
type MockSystemdData struct {
	Statuses  map[string]UnitStatus
	UnitFiles map[string][]*unit.UnitOption
}

// NewMockSystemd creates a new MockSystemd.
func NewMockSystemd() *MockSystemd {
	return &MockSystemd{
		Data: &MockSystemdData{
			Statuses:  make(map[string]UnitStatus),
			UnitFiles: make(map[string][]*unit.UnitOption),
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
	s.Data.UnitFiles[args.Name] = args.UnitOptions
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

	s.ManualEnable(args.Name)
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
	list := make([]UnitStatus, 0, len(s.Data.Statuses))
	for _, status := range s.Data.Statuses {
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

	delete(s.Data.Statuses, args.Name)
	delete(s.Data.UnitFiles, args.Name)
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

// ClearData clears all data out of the mock provider.
func (s *MockSystemd) ClearData() {
	s.Data = &MockSystemdData{
		Statuses:  make(map[string]UnitStatus),
		UnitFiles: make(map[string][]*unit.UnitOption),
	}
}

// ManualCreate directly creates a service in the mock data, optionally enabled.
func (s *MockSystemd) ManualCreate(args CreateArgs, enable bool) {
	s.Data.UnitFiles[args.Name] = args.UnitOptions
	if enable {
		s.ManualEnable(args.Name)
	}
}

// ManualEnable directly enables a service in the mock data.
func (s *MockSystemd) ManualEnable(name string) {
	unit, ok := s.Data.UnitFiles[name]
	if !ok {
		return
	}

	s.Data.Statuses[name] = UnitStatus{
		UnitStatus: dbus.UnitStatus{
			Name:        name,
			LoadState:   "Loaded",
			ActiveState: "Active",
		},
		Uptime:             time.Minute,
		UnitProperties:     make(map[string]interface{}),
		UnitTypeProperties: make(map[string]interface{}),
	}
	var env []string
	for _, unitOption := range unit {
		key := unitOption.Name
		var value interface{}
		// since not everything is a string when looked up, do conversions of
		// properties we know things care about
		switch unitOption.Name {
		case "ExecStart":
			value = [][]interface{}{{"", strings.Split(unitOption.Value, " ")}}
		case "Description":
			value = unitOption.Value
			//s.Data.Statuses[name].Description = unitOption.Value
			tmp := s.Data.Statuses[name]
			tmp.Description = unitOption.Value
			s.Data.Statuses[name] = tmp
		case "Environment":
			env = append(env, unitOption.Value)
		default:
			value = unitOption.Value
		}
		// fudge it here for simplicity
		s.Data.Statuses[name].UnitProperties[key] = value
		s.Data.Statuses[name].UnitTypeProperties[key] = value
	}
	s.Data.Statuses[name].UnitTypeProperties["Environment"] = env
}

// ManualGet directly retrieves a services from the mock data.
func (s *MockSystemd) ManualGet(name string) *UnitStatus {
	u, ok := s.Data.Statuses[name]
	if ok {
		return &u
	}
	return nil
}
