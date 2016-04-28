# zfs

[![zfs](https://godoc.org/github.com/cerana/cerana/zfs?status.svg)](https://godoc.org/github.com/cerana/cerana/zfs)



## Usage

```go
const (
	DatasetFilesystem = "filesystem"
	DatasetSnapshot   = "snapshot"
	DatasetVolume     = "volume"
)
```
ZFS Dataset Types

#### func  Datasets

```go
func Datasets(name string, dsTypes []string) ([]*Dataset, error)
```
Datasets retrieves a list of datasets of specified types. If types are not
specified, all types will be returned.

#### func  Exists

```go
func Exists(name string) (bool, error)
```
Exists determines whether a dataset exists or not.

#### func  Receive

```go
func Receive(stream io.Reader, name string) error
```
Receive creates a snapshot from a zfs send stream.

#### type By

```go
type By func(p1, p2 *Dataset) bool
```

By is the type of a "less" function that defines the ordering of its Dataset
arguments.

#### func (By) Sort

```go
func (by By) Sort(datasets []*Dataset)
```
Sort is a method on the function type, By, that sorts the argument slice
according to the function.

#### type DMUObjsetStats

```go
type DMUObjsetStats struct {
	CreationTxg  uint64 `nv:"dds_creation_txg"`
	GUID         uint64 `nv:"dds_guid"`
	Inconsistent bool   `nv:"dds_inconsistent"`
	IsSnapshot   bool   `nv:"dds_is_snapshot"`
	NumClones    uint64 `nv:"dds_num_clones"`
	Origin       string `nv:"dds_origin"`
	Type         string `nv:"dds_type"`
}
```

DMUObjsetStats represents zfs dataset information.

#### type Dataset

```go
type Dataset struct {
	Name           string
	Properties     *DatasetProperties
	DMUObjsetStats *DMUObjsetStats
}
```

Dataset contains information and properties for a ZFS dataset.

#### func  CreateFilesystem

```go
func CreateFilesystem(name string, properties map[string]interface{}) (*Dataset, error)
```
CreateFilesystem creates a new filesystem.

#### func  CreateVolume

```go
func CreateVolume(name string, size uint64, properties map[string]interface{}) (*Dataset, error)
```
CreateVolume creates a new volume.

#### func  GetDataset

```go
func GetDataset(name string) (*Dataset, error)
```
GetDataset retrieves a single dataset.

#### func (*Dataset) Children

```go
func (d *Dataset) Children(depth uint64) ([]*Dataset, error)
```
Children returns a list of children of the dataset.

#### func (*Dataset) Clone

```go
func (d *Dataset) Clone(name string, properties map[string]interface{}) (*Dataset, error)
```
Clone clones a snapshot and returns a clone dataset.

#### func (*Dataset) Destroy

```go
func (d *Dataset) Destroy(opts *DestroyOptions) error
```
Destroy destroys a zfs dataset, optionally recursive for descendants and clones.
Note that recursive destroys are not an atomic operation.

#### func (*Dataset) Diff

```go
func (d *Dataset) Diff(name string)
```
Diff returns changes between a snapshot and the given dataset. Currently a stub.

#### func (*Dataset) Holds

```go
func (d *Dataset) Holds() ([]string, error)
```
Holds returns a list of user holds on the dataset.

#### func (*Dataset) Mount

```go
func (d *Dataset) Mount(overlay bool, options []string) error
```
Mount mounts the dataset.

#### func (*Dataset) Mountpoint

```go
func (d *Dataset) Mountpoint() string
```
Mountpoint returns the mountpoint of the dataset. It is based off of the dataset
mountpoint property joined to the dataset name with the mountpointsource
property trimmed from the name.

#### func (*Dataset) Pool

```go
func (d *Dataset) Pool() string
```
Pool returns the zfs pool the dataset belongs to.

#### func (*Dataset) Rename

```go
func (d *Dataset) Rename(newName string, recursive bool) (string, error)
```
Rename renames the dataset, returning the failed name on error.

#### func (*Dataset) Rollback

```go
func (d *Dataset) Rollback(destroyMoreRecent bool) error
```
Rollback rolls back a dataset to a previous snapshot.

#### func (*Dataset) Send

```go
func (d *Dataset) Send(output io.Writer) error
```
Send sends a stream of a snapshot to the writer.

#### func (*Dataset) SetProperty

```go
func (d *Dataset) SetProperty(name string, value interface{}) error
```
SetProperty sets the value of a property of the dataset. Currently a stub.

#### func (*Dataset) Snapshot

```go
func (d *Dataset) Snapshot(name string, recursive bool) error
```
Snapshot creates a new snapshot of the dataset.

#### func (*Dataset) Snapshots

```go
func (d *Dataset) Snapshots() ([]*Dataset, error)
```
Snapshots returns a list of snapshots of the dataset.

#### func (*Dataset) Unmount

```go
func (d *Dataset) Unmount(force bool) error
```
Unmount unmounts the dataset.

#### type DatasetProperties

```go
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
```

DatasetProperties are properties of a ZFS dataset. Some properties may be
modified from the values returned by zfs.

#### type DestroyOptions

```go
type DestroyOptions struct {
	Recursive       bool
	RecursiveClones bool
	ForceUnmount    bool
	Defer           bool
}
```

DestroyOptions are used to determine the behavior when destroying a dataset.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
