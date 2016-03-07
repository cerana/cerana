package gozfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/mistifyio/gozfs/nv"
)

// ZFS Dataset Types
const (
	DatasetFilesystem = "filesystem"
	DatasetSnapshot   = "snapshot"
	DatasetVolume     = "volume"
)

// Dataset is a ZFS dataset containing a simplified set of information.
type Dataset struct {
	Name          string
	Origin        string
	Used          uint64
	Avail         uint64
	Mountpoint    string
	Compression   string
	Type          string
	Written       uint64
	Volsize       uint64
	Usedbydataset uint64
	Logicalused   uint64
	Quota         uint64
	ds            *ds
}

// By is the type of a "less" function that defines the ordering of its Dataset arguments.
type By func(p1, p2 *Dataset) bool

// Sort is a method on the function type, By, that sorts the argument slice according to the function.
func (by By) Sort(datasets []*Dataset) {
	ds := &datasetSorter{
		datasets: datasets,
		by:       by, // The Sort method's receiver is the function (closure) that defines the sort order.
	}
	sort.Sort(ds)
}

// datasetSorter joins a By function and a slice of Datasets to be sorted.
type datasetSorter struct {
	datasets []*Dataset
	by       func(p1, p2 *Dataset) bool // Closure used in the Less method.
}

// Len is part of sort.Interface.
func (s *datasetSorter) Len() int {
	return len(s.datasets)
}

// Swap is part of sort.Interface.
func (s *datasetSorter) Swap(i, j int) {
	s.datasets[i], s.datasets[j] = s.datasets[j], s.datasets[i]
}

// Less is part of sort.Interface. It is implemented by calling the "by" closure in the sorter.
func (s *datasetSorter) Less(i, j int) bool {
	return s.by(s.datasets[i], s.datasets[j])
}

type ds struct {
	DMUObjsetStats *dmuObjsetStats `nv:"dmu_objset_stats"`
	Name           string          `nv:"name"`
	Properties     *dsProperties   `nv:"properties"`
}

type dmuObjsetStats struct {
	CreationTxg  uint64 `nv:"dds_creation_txg"`
	GUID         uint64 `nv:"dds_guid"`
	Inconsistent bool   `nv:"dds_inconsistent"`
	IsSnapshot   bool   `nv:"dds_is_snapshot"`
	NumClones    uint64 `nv:"dds_num_clonse"`
	Origin       string `nv:"dds_origin"`
	Type         string `nv:"dds_type"`
}

type dsProperties struct {
	Available            propUint64           `nv:"available"`
	Clones               propClones           `nv:"clones"`
	Compression          propStringWithSource `nv:"compression"`
	CompressRatio        propUint64           `nv:"compressratio"`
	CreateTxg            propUint64           `nv:"createtxg"`
	Creation             propUint64           `nv:"creation"`
	DeferDestroy         propUint64           `nv:"defer_destroy"`
	GUID                 propUint64           `nv:"guid"`
	LogicalReferenced    propUint64           `nv:"logicalreferenced"`
	LogicalUsed          propUint64           `nv:"logicalused"`
	Mountpoint           propStringWithSource `nv:"mountpoint"`
	ObjsetID             propUint64           `nv:"objsetid"`
	Origin               propString           `nv:"origin"`
	Quota                propUint64WithSource `nv:"quota"`
	RefCompressRatio     propUint64           `nv:"refcompressratio"`
	RefQuota             propUint64WithSource `nv:"refquota"`
	RefReservation       propUint64WithSource `nv:"refreservation"`
	Referenced           propUint64           `nv:"referenced"`
	Reservation          propUint64WithSource `nv:"reservation"`
	Type                 propUint64           `nv:"type"`
	Unique               propUint64           `nv:"unique"`
	Used                 propUint64           `nv:"used"`
	UsedByChildren       propUint64           `nv:"usedbychildren"`
	UsedByDataset        propUint64           `nv:"usedbydataset"`
	UsedByRefReservation propUint64           `nv:"usedbyrefreservation"`
	UsedBySnapshots      propUint64           `nv:"usedbysnapshots"`
	UserAccounting       propUint64           `nv:"useraccounting"`
	UserRefs             propUint64           `nv:"userrefs"`
	Volsize              propUint64           `nv:"volsize"`
	VolBlockSize         propUint64           `nv:"volblocksize"`
	Written              propUint64           `nv:"written"`
}

var dsPropertyIndexes map[string]int

type dsProperty interface {
	value() interface{}
}

type propClones struct {
	Value map[string]nv.Boolean `nv:"value"`
}

func (p propClones) value() []string {
	clones := make([]string, len(p.Value))
	i := 0
	for clone := range p.Value {
		clones[i] = clone
		i++
	}
	return clones
}

type propUint64 struct {
	Value uint64 `nv:"value"`
}

func (p propUint64) value() uint64 {
	return p.Value
}

type propUint64WithSource struct {
	Source string `nv:"source"`
	Value  uint64 `nv:"value"`
}

func (p propUint64WithSource) value() uint64 {
	return p.Value
}

type propString struct {
	Value string `nv:"value"`
}

func (p propString) value() string {
	return p.Value
}

type propStringWithSource struct {
	Source string `nv:"source"`
	Value  string `nv:"value"`
}

func (p propStringWithSource) value() string {
	return p.Value
}

func dsToDataset(in *ds) *Dataset {
	var dsType string
	if in.DMUObjsetStats.IsSnapshot {
		dsType = DatasetSnapshot
	} else if dmuType(in.Properties.Type.Value) == dmuTypes["zvol"] {
		dsType = DatasetVolume
	} else {
		dsType = DatasetFilesystem
	}

	compression := in.Properties.Compression.Value
	if compression == "" {
		compression = "off"
	}

	mountpoint := in.Properties.Mountpoint.Value
	if mountpoint == "" && dsType != DatasetSnapshot {
		mountpoint = fmt.Sprintf("/%s", in.Name)
	}

	return &Dataset{
		Name:          in.Name,
		Origin:        in.Properties.Origin.Value,
		Used:          in.Properties.Used.Value,
		Avail:         in.Properties.Available.Value,
		Mountpoint:    mountpoint,
		Compression:   compression,
		Type:          dsType,
		Written:       in.Properties.Available.Value,
		Volsize:       in.Properties.Volsize.Value,
		Usedbydataset: in.Properties.UsedByDataset.Value,
		Logicalused:   in.Properties.LogicalUsed.Value,
		Quota:         in.Properties.Quota.Value,
		ds:            in,
	}
}

func getDatasets(name, dsType string, recurse bool, depth uint64) ([]*Dataset, error) {
	types := map[string]bool{
		dsType: true,
	}

	dss, err := list(name, types, recurse, depth)
	if err != nil {
		return nil, err
	}

	datasets := make([]*Dataset, len(dss))
	for i, ds := range dss {
		datasets[i] = dsToDataset(ds)
	}

	byName := func(d1, d2 *Dataset) bool {
		return d1.Name < d2.Name
	}
	By(byName).Sort(datasets)

	return datasets, nil
}

// Datasets retrieves a list of all datasets, regardless of type.
func Datasets(name string) ([]*Dataset, error) {
	return getDatasets(name, "all", true, 0)
}

// Filesystems retrieves a list of all filesystems.
func Filesystems(name string) ([]*Dataset, error) {
	return getDatasets(name, DatasetFilesystem, true, 0)
}

// Snapshots retrieves a list of all snapshots.
func Snapshots(name string) ([]*Dataset, error) {
	return getDatasets(name, DatasetSnapshot, true, 0)
}

// Volumes retrieves a list of all volumes.
func Volumes(name string) ([]*Dataset, error) {
	return getDatasets(name, DatasetVolume, true, 0)
}

// GetDataset retrieves a single dataset.
func GetDataset(name string) (*Dataset, error) {
	datasets, err := getDatasets(name, "all", false, 0)
	if err != nil {
		return nil, err
	}
	if len(datasets) != 1 {
		return nil, fmt.Errorf("expected 1 dataset, got %d", len(datasets))
	}
	return datasets[0], nil
}

func createDataset(name string, createType dmuType, properties map[string]interface{}) (*Dataset, error) {
	if err := create(name, createType, properties); err != nil {
		return nil, err
	}

	return GetDataset(name)
}

// CreateFilesystem creates a new filesystem.
func CreateFilesystem(name string, properties map[string]interface{}) (*Dataset, error) {
	return createDataset(name, dmuZFS, properties)
}

// CreateVolume creates a new volume.
func CreateVolume(name string, size uint64, properties map[string]interface{}) (*Dataset, error) {
	if size != 0 {
		if properties == nil {
			properties = make(map[string]interface{})
		}
		properties["volsize"] = size
	}
	return createDataset(name, dmuZVOL, properties)
}

// ReceiveSnapshot creates a snapshot from a zfs send stream. Currently a stub.
func ReceiveSnapshot(input io.Reader, name string) (*Dataset, error) {
	// TODO: Implement when we have a zfs_receive
	return nil, errors.New("zfs receive not yet implemented")
}

// Children returns a list of children of the dataset.
func (d *Dataset) Children(depth uint64) ([]*Dataset, error) {
	datasets, err := getDatasets(d.Name, "all", true, depth)
	if err != nil {
		return nil, err
	}
	return datasets[1:], nil
}

// Clone clones a snapshot and returns a clone dataset.
func (d *Dataset) Clone(name string, properties map[string]interface{}) (*Dataset, error) {
	if err := clone(name, d.Name, properties); err != nil {
		return nil, err
	}
	return GetDataset(name)
}

// DestroyOptions are used to determine the behavior when destroying a dataset.
type DestroyOptions struct {
	Recursive       bool
	RecursiveClones bool
	ForceUnmount    bool
	Defer           bool
}

// Destroy destroys a zfs dataset, optionally recursive for descendants and clones.
// Note that recursive destroys are not an atomic operation.
func (d *Dataset) Destroy(opts *DestroyOptions) error {
	// Recurse
	if opts.Recursive {
		children, err := d.Children(1)
		if err != nil {
			return err
		}
		for _, child := range children {
			if err := child.Destroy(opts); err != nil {
				return err
			}
		}
	}

	// Recurse Clones
	if opts.RecursiveClones {
		for cloneName := range d.ds.Properties.Clones.Value {
			clone, err := GetDataset(cloneName)
			if err != nil {
				return err
			}
			if err := clone.Destroy(opts); err != nil {
				return err
			}
		}
	}

	// Unmount this dataset
	// TODO: Implement when we have a zfs_unmount

	// Destroy this dataset
	return destroy(d.Name, opts.Defer)
}

// Diff returns changes between a snapshot and the given dataset. Currently a stub.
func (d *Dataset) Diff(name string) {
	// TODO: Implement when we have a zfs_diff
}

// GetProperty returns the current value of a property from the dataset.
func (d *Dataset) GetProperty(name string) (interface{}, error) {
	propertyIndex, ok := dsPropertyIndexes[strings.ToLower(name)]
	dV := reflect.ValueOf(d.ds.Properties)
	if !ok {
		return nil, errors.New("not a valid property name")
	}
	property := reflect.Indirect(dV).Field(propertyIndex).Interface().(dsProperty)
	return property.value(), nil
}

// SetProperty sets the value of a property of the dataset. Currently a stub.
func (d *Dataset) SetProperty(name string, value interface{}) error {
	// TODO: Implement when we have a zfs_set_property
	return errors.New("zfs set property not implemented yet")
}

// Rollback rolls back a dataset to a previous snapshot.
func (d *Dataset) Rollback(destroyMoreRecent bool) error {
	// Name of dataset the snapshot belongs to
	dsName := strings.Split(d.Name, "@")[0]

	// Get all of the dataset's snapshots
	snapshots, err := getDatasets(dsName, DatasetSnapshot, true, 1)
	if err != nil {
		return err
	}

	// Order snapshots from oldest to newest
	creation := func(d1, d2 *Dataset) bool {
		return d1.ds.Properties.Creation.Value < d2.ds.Properties.Creation.Value
	}
	By(creation).Sort(snapshots)

	// Destroy any snapshots newer than the target
	found := false
	for _, snapshot := range snapshots {
		// Ignore this snapshot and all older
		if !found {
			if snapshot.Name == d.Name {
				found = true
			}
			continue
		}

		// Only destroy if the flag is explicitly set
		if !destroyMoreRecent {
			return errors.New("not most recent snapshot")
		}

		opts := &DestroyOptions{
			Recursive:       true,
			RecursiveClones: true,
		}
		if err := snapshot.Destroy(opts); err != nil {
			return err
		}
	}

	// Rollback to the target snapshot, which is now the most recent
	_, err = rollback(dsName)
	return err
}

// fdCloser is the interface for zfs send output destination.
// send requires a file descriptor.
// Close will only be called on internally created instances used to connect non-implementing io.Writers to send.
type fdCloser interface {
	Fd() uintptr
	Close() error
}

// Send sends a stream of a snapshot to the writer.
func (d *Dataset) Send(output io.Writer) error {
	done := make(chan error, 1)

	outputFdCloser, isFdCloser := output.(fdCloser)
	// Connect a file descriptor to the output Writer
	if !isFdCloser {
		r, w, err := os.Pipe()
		if err != nil {
			return err
		}
		defer w.Close()
		defer r.Close()

		outputFdCloser = w

		// Copy data to the output Writer
		go func() {
			_, err := io.Copy(output, r)
			done <- err
		}()
	} else {
		done <- nil
	}

	if err := send(d.Name, outputFdCloser.Fd(), "", false, false); err != nil {
		return err
	}
	if !isFdCloser {
		outputFdCloser.Close()
	}
	return <-done
}

// Snapshot creates a new snapshot of the dataset.
func (d *Dataset) Snapshot(name string, recursive bool) error {
	snapNames := []string{fmt.Sprintf("%s@%s", d.Name, name)}
	props := map[string]string{}
	if _, err := snapshot(d.Pool(), snapNames, props); err != nil {
		return err
	}
	if recursive {
		children, err := d.Children(1)
		if err != nil {
			return err
		}
		for _, child := range children {
			// Can't snapshot a snapshot
			if child.Type == DatasetSnapshot {
				continue
			}
			if err := child.Snapshot(name, recursive); err != nil {
				return err
			}
		}
	}
	return nil
}

// Snapshots returns a list of snapshots of the dataset.
func (d *Dataset) Snapshots() ([]*Dataset, error) {
	return Snapshots(d.Name)
}

// Pool returns the zfs pool the dataset belongs to.
func (d *Dataset) Pool() string {
	return strings.Split(d.Name, "/")[0]
}

// Holds returns a list of user holds on the dataset.
func (d *Dataset) Holds() ([]string, error) {
	return holds(d.Name)
}

// Rename renames the dataset, returning the failed name on error.
func (d *Dataset) Rename(newName string, recursive bool) (string, error) {
	if failedName, err := rename(d.Name, newName, recursive); err != nil {
		return failedName, err
	}

	d.Name = newName
	return "", nil
}

func init() {
	dsPropertyIndexes = make(map[string]int)
	dsPropertiesT := reflect.TypeOf(dsProperties{})
	for i := 0; i < dsPropertiesT.NumField(); i++ {
		field := dsPropertiesT.Field(i)
		name := field.Name
		tags := strings.Split(field.Tag.Get("nv"), ",")
		if len(tags) > 0 && tags[0] != "" {
			name = tags[0]
		}
		dsPropertyIndexes[strings.ToLower(name)] = i
	}
}
