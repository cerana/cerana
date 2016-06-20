package zfs

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/zfs"
)

// MockZFS is a mock ZFS provider.
type MockZFS struct {
	config  *provider.Config
	tracker *acomm.Tracker
	Data    *MockZFSData
}

// MockZFSData is the in-memory data structure for the MockZFS.
type MockZFSData struct {
	Datasets map[string]*Dataset
	Data     map[string][]byte
}

// NewMockZFS creates a new instance of MockZFS.
func NewMockZFS(config *provider.Config, tracker *acomm.Tracker) *MockZFS {
	return &MockZFS{
		config:  config,
		tracker: tracker,
		Data: &MockZFSData{
			Datasets: make(map[string]*Dataset),
			Data:     make(map[string][]byte),
		},
	}
}

// RegisterTasks registers all MockZFS tasks.
func (z *MockZFS) RegisterTasks(server *provider.Server) {
	server.RegisterTask("zfs-clone", z.Clone)
	server.RegisterTask("zfs-create", z.Create)
	server.RegisterTask("zfs-destroy", z.Destroy)
	server.RegisterTask("zfs-exists", z.Exists)
	server.RegisterTask("zfs-get", z.Get)
	server.RegisterTask("zfs-holds", z.Holds)
	server.RegisterTask("zfs-list", z.List)
	server.RegisterTask("zfs-mount", z.Mount)
	server.RegisterTask("zfs-receive", z.Receive)
	server.RegisterTask("zfs-rename", z.Rename)
	server.RegisterTask("zfs-rollback", z.Rollback)
	server.RegisterTask("zfs-send", z.Send)
	server.RegisterTask("zfs-snapshot", z.Snapshot)
	server.RegisterTask("zfs-unmount", z.Unmount)
}

// Clone clones a mock dataset.
func (z *MockZFS) Clone(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CloneArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if args.Origin == "" {
		return nil, nil, errors.New("missing arg: origin")
	}
	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}

	origin, ok := z.Data.Datasets[args.Origin]
	if !ok {
		return nil, nil, errors.New("dataset not found")
	}
	*z.Data.Datasets[args.Name] = *origin
	z.Data.Datasets[args.Name].Name = args.Name
	z.Data.Datasets[args.Name].Properties.Origin = args.Origin
	return &DatasetResult{z.Data.Datasets[args.Name]}, nil, nil
}

// Create creats a mock dataset.
func (z *MockZFS) Create(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CreateArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if err := fixPropertyTypesFromJSON(args.Properties); err != nil {
		return nil, nil, err
	}
	_, ok := z.Data.Datasets[args.Name]
	if ok {
		return nil, nil, errors.New("dataset not found")
	}

	z.Data.Datasets[args.Name] = &Dataset{
		Name: args.Name,
		Properties: &zfs.DatasetProperties{
			Type:    args.Type,
			Volsize: args.Volsize,
		},
	}
	return &DatasetResult{z.Data.Datasets[args.Name]}, nil, nil
}

// Destroy destroys a mock dataset.
func (z *MockZFS) Destroy(req *acomm.Request) (interface{}, *url.URL, error) {
	var args DestroyArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}
	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := z.Data.Datasets[args.Name]; !ok {
		return nil, nil, errors.New("dataset not found")
	}
	delete(z.Data.Datasets, args.Name)
	delete(z.Data.Data, args.Name)
	return nil, nil, nil
}

// Exists checks whether a mock dataset exists.
func (z *MockZFS) Exists(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	_, ok := z.Data.Datasets[args.Name]
	return &ExistsResult{ok}, nil, nil
}

// Get retrieves a mock dataset.
func (z *MockZFS) Get(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	dataset, ok := z.Data.Datasets[args.Name]
	if !ok {
		return nil, nil, errors.New("dataset not found")
	}
	return &DatasetResult{dataset}, nil, nil
}

// Holds retrieves a mock dataset's holds.
func (z *MockZFS) Holds(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	if _, ok := z.Data.Datasets[args.Name]; !ok {
		return nil, nil, errors.New("dataset not found")
	}

	return &HoldsResult{[]string{}}, nil, nil
}

// List returns all mock datasets.
func (z *MockZFS) List(req *acomm.Request) (interface{}, *url.URL, error) {
	var args ListArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if len(args.Types) == 0 {
		args.Types = []string{"all"}
	}
	datasets := make([]*Dataset, 0, len(z.Data.Datasets))
	for _, dataset := range z.Data.Datasets {
		for _, t := range args.Types {
			if t == dataset.Properties.Type || t == "all" {
				datasets = append(datasets, dataset)
				break
			}
		}
	}

	return &ListResult{datasets}, nil, nil
}

// Mount mounts a mock dataset.
func (z *MockZFS) Mount(req *acomm.Request) (interface{}, *url.URL, error) {
	var args MountArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if _, ok := z.Data.Datasets[args.Name]; !ok {
		return nil, nil, errors.New("dataset not found")
	}

	return nil, nil, nil
}

// Receive receives mock dataset data and creates a mock dataset.
func (z *MockZFS) Receive(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if req.StreamURL == nil {
		return nil, nil, errors.New("missing request stream-url")
	}

	r, w := io.Pipe()
	go func() {
		defer logrusx.LogReturnedErr(w.Close, nil, "failed to close writer")
		logrusx.LogReturnedErr(func() error {
			return acomm.Stream(w, req.StreamURL)
		}, logrus.Fields{"streamURL": req.StreamURL}, "failed to stream")
	}()

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	z.Data.Data[args.Name] = data
	z.Data.Datasets[args.Name] = &Dataset{Name: args.Name}
	return nil, nil, nil
}

// Rename renames a mock dataset.
func (z *MockZFS) Rename(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RenameArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if args.Origin == "" {
		return nil, nil, errors.New("missing arg: origin")
	}

	origin, ok := z.Data.Datasets[args.Origin]
	if !ok {
		return nil, nil, errors.New("dataset not found")
	}
	if _, ok = z.Data.Datasets[args.Name]; ok {
		return nil, nil, errors.New("dataset already exists")
	}

	z.Data.Datasets[args.Name] = origin
	delete(z.Data.Datasets, args.Origin)
	if data, ok := z.Data.Data[args.Origin]; ok {
		z.Data.Data[args.Name] = data
		delete(z.Data.Data, args.Origin)
	}
	return nil, nil, nil
}

// Rollback rolls back to a mock dataset.
func (z *MockZFS) Rollback(req *acomm.Request) (interface{}, *url.URL, error) {
	var args RollbackArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if _, ok := z.Data.Datasets[args.Name]; !ok {
		return nil, nil, errors.New("dataset not found")
	}
	return nil, nil, nil
}

// Send sends mock dataset data.
func (z *MockZFS) Send(req *acomm.Request) (interface{}, *url.URL, error) {
	var args CommonArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}

	data, ok := z.Data.Data[args.Name]
	if !ok {
		return nil, nil, errors.New("dataset not found")
	}

	reader := bytes.NewReader(data)

	addr, err := z.tracker.NewStreamUnix(z.config.StreamDir("zfs-send"), ioutil.NopCloser(reader))
	if err != nil {
		return nil, nil, err
	}

	return nil, addr, nil
}

// Snapshot snapshots a mock dataset.
func (z *MockZFS) Snapshot(req *acomm.Request) (interface{}, *url.URL, error) {
	var args SnapshotArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if args.SnapName == "" {
		return nil, nil, errors.New("missing arg: snapname")
	}

	dataset, ok := z.Data.Datasets[args.Name]
	if !ok {
		return nil, nil, errors.New("dataset not found")
	}

	snapName := args.Name + "@" + args.SnapName
	*z.Data.Datasets[snapName] = *dataset
	z.Data.Datasets[snapName].Name = snapName
	z.Data.Datasets[snapName].Properties.Type = "snapshot"

	return nil, nil, nil
}

// Unmount unmounts a mock dataset.
func (z *MockZFS) Unmount(req *acomm.Request) (interface{}, *url.URL, error) {
	var args UnmountArgs
	if err := req.UnmarshalArgs(&args); err != nil {
		return nil, nil, err
	}

	if args.Name == "" {
		return nil, nil, errors.New("missing arg: name")
	}
	if _, ok := z.Data.Datasets[args.Name]; !ok {
		return nil, nil, errors.New("dataset not found")
	}

	return nil, nil, nil
}
