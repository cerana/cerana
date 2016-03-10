package gozfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

// Dataset contains information and properties for a ZFS dataset.
type Dataset struct {
	Name           string
	Properties     *DatasetProperties
	DMUObjsetStats *DMUObjsetStats
}

// DatasetProperties are properties of a ZFS dataset. Some properties may be
// modified from the values returned by zfs.
type DatasetProperties struct {
	Available             uint64
	CaseSensitivity       uint64
	CaseSensitivitySource string
	Clones                []string
	CompressRatio         uint64
	Compression           string // Empty value is replaced with "off"
	CompressionSource     string
	CreateTxg             uint64
	Creation              uint64
	DeferDestroy          uint64
	GUID                  uint64
	LogicalReferenced     uint64
	LogicalUsed           uint64
	Mountpoint            string
	MountpointSource      string
	Normalization         uint64
	NormalizationSource   string
	ObjsetID              uint64
	Origin                string
	PrevSnap              string
	PrevSnapSource        string
	Quota                 uint64
	QuotaSource           string
	RefCompressRatio      uint64
	RefQuota              uint64
	RefQuotaSource        string
	RefReservation        uint64
	RefReservationSource  string
	Referenced            uint64
	Reservation           uint64
	ReservationSource     string
	Type                  string // Int type mapped to string type
	UTF8Only              bool
	UTF8OnlySource        string
	Unique                uint64
	Used                  uint64
	UsedByChildren        uint64
	UsedByDataset         uint64
	UsedByRefReservation  uint64
	UsedBySnapshots       uint64
	UserAccounting        uint64
	UserDefined           map[string]string
	UserRefs              uint64
	Version               uint64
	VolBlockSize          uint64
	VolBlockSizeSource    string
	Volsize               uint64
	Written               uint64
}

// DMUObjsetStats represents zfs dataset information.
type DMUObjsetStats struct {
	CreationTxg  uint64 `nv:"dds_creation_txg"`
	GUID         uint64 `nv:"dds_guid"`
	Inconsistent bool   `nv:"dds_inconsistent"`
	IsSnapshot   bool   `nv:"dds_is_snapshot"`
	NumClones    uint64 `nv:"dds_num_clones"`
	Origin       string `nv:"dds_origin"`
	Type         string `nv:"dds_type"`
}

// dataset is an intermediate struct for unmarshalling a zfs dataset nvlist.
type dataset struct {
	DMUObjsetStats *DMUObjsetStats    `nv:"dmu_objset_stats"`
	Name           string             `nv:"name"`
	Properties     *datasetProperties `nv:"properties"`
}

// datasetProperties is an intermediate struct for unmarshalling zfs dataset
// properties nvlist.
type datasetProperties struct {
	Available            propUint64                      `nv:"available"`
	CaseSensitivity      propUint64WithSource            `nv:"casesensitivity"`
	Clones               propClones                      `nv:"clones"`
	CompressRatio        propUint64                      `nv:"compressratio"`
	Compression          propStringWithSource            `nv:"compression"`
	CreateTxg            propUint64                      `nv:"createtxg"`
	Creation             propUint64                      `nv:"creation"`
	DeferDestroy         propUint64                      `nv:"defer_destroy"`
	GUID                 propUint64                      `nv:"guid"`
	LogicalReferenced    propUint64                      `nv:"logicalreferenced"`
	LogicalUsed          propUint64                      `nv:"logicalused"`
	Mountpoint           propStringWithSource            `nv:"mountpoint"`
	Normalization        propUint64WithSource            `nv:"normalization"`
	ObjsetID             propUint64                      `nv:"objsetid"`
	Origin               propString                      `nv:"origin"`
	PrevSnap             propStringWithSource            `nv:"prevsnap"`
	Quota                propUint64WithSource            `nv:"quota"`
	RefCompressRatio     propUint64                      `nv:"refcompressratio"`
	RefQuota             propUint64WithSource            `nv:"refquota"`
	RefReservation       propUint64WithSource            `nv:"refreservation"`
	Referenced           propUint64                      `nv:"referenced"`
	Reservation          propUint64WithSource            `nv:"reservation"`
	Type                 propUint64                      `nv:"type"`
	UTF8Only             propUint64WithSource            `nv:"utf8only"`
	Unique               propUint64                      `nv:"unique"`
	Used                 propUint64                      `nv:"used"`
	UsedByChildren       propUint64                      `nv:"usedbychildren"`
	UsedByDataset        propUint64                      `nv:"usedbydataset"`
	UsedByRefReservation propUint64                      `nv:"usedbyrefreservation"`
	UsedBySnapshots      propUint64                      `nv:"usedbysnapshots"`
	UserAccounting       propUint64                      `nv:"useraccounting"`
	UserDefined          map[string]propStringWithSource `nv:",extra"`
	UserRefs             propUint64                      `nv:"userrefs"`
	Version              propUint64                      `nv:"version"`
	VolBlockSize         propUint64WithSource            `nv:"volblocksize"`
	Volsize              propUint64                      `nv:"volsize"`
	Written              propUint64                      `nv:"written"`
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

// toDataset returns a new Dataset based on the intermediate dataset.
func (ds *dataset) toDataset() *Dataset {
	var dsType string
	if ds.DMUObjsetStats.IsSnapshot {
		dsType = DatasetSnapshot
	} else if dmuType(ds.Properties.Type.Value) == dmuTypes["zvol"] {
		dsType = DatasetVolume
	} else {
		dsType = DatasetFilesystem
	}

	compression := ds.Properties.Compression.Value
	if compression == "" {
		compression = "off"
	}

	userDefined := make(map[string]string)
	for key, value := range ds.Properties.UserDefined {
		userDefined[key] = value.Value
		userDefined[key+"Source"] = value.Source
	}

	return &Dataset{
		Name: ds.Name,
		Properties: &DatasetProperties{
			Available:             ds.Properties.Available.Value,
			CaseSensitivity:       ds.Properties.CaseSensitivity.Value,
			CaseSensitivitySource: ds.Properties.CaseSensitivity.Source,
			Clones:                ds.Properties.Clones.value(),
			CompressRatio:         ds.Properties.CompressRatio.Value,
			Compression:           compression,
			CompressionSource:     ds.Properties.Compression.Source,
			CreateTxg:             ds.Properties.CreateTxg.Value,
			Creation:              ds.Properties.Creation.Value,
			DeferDestroy:          ds.Properties.DeferDestroy.Value,
			GUID:                  ds.Properties.GUID.Value,
			LogicalReferenced:     ds.Properties.LogicalReferenced.Value,
			LogicalUsed:           ds.Properties.LogicalUsed.Value,
			Mountpoint:            ds.Properties.Mountpoint.Value,
			MountpointSource:      ds.Properties.Mountpoint.Source,
			Normalization:         ds.Properties.Normalization.Value,
			NormalizationSource:   ds.Properties.Normalization.Source,
			ObjsetID:              ds.Properties.ObjsetID.Value,
			Origin:                ds.Properties.Origin.Value,
			PrevSnap:              ds.Properties.PrevSnap.Value,
			PrevSnapSource:        ds.Properties.PrevSnap.Source,
			Quota:                 ds.Properties.Quota.Value,
			QuotaSource:           ds.Properties.Quota.Source,
			RefCompressRatio:      ds.Properties.RefCompressRatio.Value,
			RefQuota:              ds.Properties.RefQuota.Value,
			RefQuotaSource:        ds.Properties.RefQuota.Source,
			RefReservation:        ds.Properties.RefReservation.Value,
			RefReservationSource:  ds.Properties.RefReservation.Source,
			Referenced:            ds.Properties.Referenced.Value,
			Reservation:           ds.Properties.Reservation.Value,
			ReservationSource:     ds.Properties.Reservation.Source,
			Type:                  dsType,
			UTF8Only:              ds.Properties.UTF8Only.Value == 1,
			UTF8OnlySource:        ds.Properties.UTF8Only.Source,
			Unique:                ds.Properties.Unique.Value,
			Used:                  ds.Properties.Used.Value,
			UsedByChildren:        ds.Properties.UsedByChildren.Value,
			UsedByDataset:         ds.Properties.UsedByDataset.Value,
			UsedByRefReservation:  ds.Properties.UsedByRefReservation.Value,
			UsedBySnapshots:       ds.Properties.UsedBySnapshots.Value,
			UserAccounting:        ds.Properties.UserAccounting.Value,
			UserDefined:           userDefined,
			UserRefs:              ds.Properties.UserRefs.Value,
			Version:               ds.Properties.Version.Value,
			VolBlockSize:          ds.Properties.VolBlockSize.Value,
			VolBlockSizeSource:    ds.Properties.VolBlockSize.Source,
			Volsize:               ds.Properties.Volsize.Value,
			Written:               ds.Properties.Written.Value,
		},
		DMUObjsetStats: ds.DMUObjsetStats,
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
		datasets[i] = ds.toDataset()
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
		for _, cloneName := range d.Properties.Clones {
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

// SetProperty sets the value of a property of the dataset. Currently a stub.
func (d *Dataset) SetProperty(name string, value interface{}) error {
	// TODO: Implement when we have a zfs_set_property
	return errors.New("zfs set property not implemented yet")
}

// Mountpoint returns the mountpoint of the dataset. It is based off of the
// dataset mountpoint property joined to the dataset name with the
// mountpointsource property trimmed from the name.
func (d *Dataset) Mountpoint() string {
	defaultPart, _ := filepath.Rel(d.Properties.MountpointSource, d.Name)
	return filepath.Join(d.Properties.Mountpoint, defaultPart)
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
		return d1.Properties.Creation < d2.Properties.Creation
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
			if child.Properties.Type == DatasetSnapshot {
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

// Dataset Sorting

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
