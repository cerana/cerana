# zfs

[![zfs](https://godoc.org/github.com/cerana/cerana/providers/zfs?status.png)](https://godoc.org/github.com/cerana/cerana/providers/zfs)



## Usage

#### type CloneArgs

```go
type CloneArgs struct {
	Name       string                 `json:"name"`
	Origin     string                 `json:"origin"`
	Properties map[string]interface{} `json:"properties"`
}
```

CloneArgs are arguments for the Clone handler.

#### type CommonArgs

```go
type CommonArgs struct {
	Name string `json:"name"` // Name of dataset
}
```

CommonArgs are arguments that apply to all handlers.

#### type CreateArgs

```go
type CreateArgs struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // gozfs.Dataset[Filesystem,Volume]
	Volsize    uint64                 `json:"volsize"`
	Properties map[string]interface{} `json:"properties"`
}
```

CreateArgs are arguments for the Create handler.

#### type Dataset

```go
type Dataset struct {
	Name           string
	Properties     *gozfs.DatasetProperties
	DMUObjsetStats *gozfs.DMUObjsetStats
}
```

Dataset contains information and properties for a ZFS dataset. This struct is
the same as gozfs.Dataset, except all methods that interact with ZFS have been
removed. The ZFS provider should be the only place that interacts with zfs
directly.

Use this struct for datasets anywhere outside the ZFS provider.

#### func (*Dataset) Mountpoint

```go
func (d *Dataset) Mountpoint() string
```
Mountpoint returns the resolved mountpoint of the dataset.

#### type DatasetResult

```go
type DatasetResult struct {
	Dataset *Dataset `json:"dataset"`
}
```

DatasetResult is a common handler result.

#### type DestroyArgs

```go
type DestroyArgs struct {
	Name            string `json:"name"`
	Defer           bool   `json:"defer"`
	Recursive       bool   `json:"recursive"`
	RecursiveClones bool   `json:"recursive-clones"`
	ForceUnmount    bool   `json:"force-unmount"`
}
```

DestroyArgs are arguments for the Destroy handler.

#### type ExistsResult

```go
type ExistsResult struct {
	Exists bool `json:"exists"`
}
```

ExistsResult is the result data for the Exists handler.

#### type HoldsResult

```go
type HoldsResult struct {
	Holds []string `json:"holds"`
}
```

HoldsResult is the result data for the Holds handler.

#### type ListArgs

```go
type ListArgs struct {
	Name  string   `json:"name"`
	Types []string `json:"types"`
}
```

ListArgs are the args for the List handler.

#### type ListResult

```go
type ListResult struct {
	Datasets []*Dataset `json:"datasets"`
}
```

ListResult is the result data for the List handler.

#### type MountArgs

```go
type MountArgs struct {
	Name    string `json:"name"`
	Overlay bool   `json:"overlay"`
}
```

MountArgs are arguments for the Mount handler.

#### type RenameArgs

```go
type RenameArgs struct {
	Name      string `json:"name"`
	Origin    string `json:"origin"`
	Recursive bool   `json:"recursive"`
}
```

RenameArgs are arguments for the Rename handler.

#### type RollbackArgs

```go
type RollbackArgs struct {
	Name          string `json:"name"`
	DestroyRecent bool   `json:"destroy-recent"`
}
```

RollbackArgs are the arguments for the Rollback handler.

#### type SnapshotArgs

```go
type SnapshotArgs struct {
	Name      string `json:"name"`
	SnapName  string `json:"snapname"`
	Recursive bool   `json:"recursive"`
}
```

SnapshotArgs are the arguments for the Snapshot handler.

#### type UnmountArgs

```go
type UnmountArgs struct {
	Name  string `json:"name"`
	Force bool   `json:"force"`
}
```

UnmountArgs are arguments for the Unmount handler.

#### type ZFS

```go
type ZFS struct {
}
```

ZFS is a provider of zfs functionality.

#### func  New

```go
func New(config *provider.Config, tracker *acomm.Tracker) *ZFS
```
New creates a new instance of ZFS.

#### func (*ZFS) Clone

```go
func (z *ZFS) Clone(req *acomm.Request) (interface{}, *url.URL, error)
```
Clone will create a clone from a snapshot.

#### func (*ZFS) Create

```go
func (z *ZFS) Create(req *acomm.Request) (interface{}, *url.URL, error)
```
Create will create a new filesystem or volume dataset.

#### func (*ZFS) Destroy

```go
func (z *ZFS) Destroy(req *acomm.Request) (interface{}, *url.URL, error)
```
Destroy will destroy a dataset.

#### func (*ZFS) Exists

```go
func (z *ZFS) Exists(req *acomm.Request) (interface{}, *url.URL, error)
```
Exists determines whether a dataset exists or not.

#### func (*ZFS) Get

```go
func (z *ZFS) Get(req *acomm.Request) (interface{}, *url.URL, error)
```
Get returns information about a dataset.

#### func (*ZFS) Holds

```go
func (z *ZFS) Holds(req *acomm.Request) (interface{}, *url.URL, error)
```
Holds retrieves a list of user holds on the specified snapshot.

#### func (*ZFS) List

```go
func (z *ZFS) List(req *acomm.Request) (interface{}, *url.URL, error)
```
List returns a list of filesystems, volumes, snapshots, and bookmarks

#### func (*ZFS) Mount

```go
func (z *ZFS) Mount(req *acomm.Request) (interface{}, *url.URL, error)
```
Mount mounts a zfs filesystem.

#### func (*ZFS) Receive

```go
func (z *ZFS) Receive(req *acomm.Request) (interface{}, *url.URL, error)
```
Receive creates a new snapshot from a zfs stream. If it a full stream, then a
new filesystem or volume is created as well.

#### func (*ZFS) RegisterTasks

```go
func (z *ZFS) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of ZFS's task handlers with the server.

#### func (*ZFS) Rename

```go
func (z *ZFS) Rename(req *acomm.Request) (interface{}, *url.URL, error)
```
Rename will create a rename from a snapshot.

#### func (*ZFS) Rollback

```go
func (z *ZFS) Rollback(req *acomm.Request) (interface{}, *url.URL, error)
```
Rollback rolls a filesystem or volume back to a given snapshot.

#### func (*ZFS) Send

```go
func (z *ZFS) Send(req *acomm.Request) (interface{}, *url.URL, error)
```
Send returns information about a dataset.

#### func (*ZFS) Snapshot

```go
func (z *ZFS) Snapshot(req *acomm.Request) (interface{}, *url.URL, error)
```
Snapshot creates a snapshot of a filesystem or volume.

#### func (*ZFS) Unmount

```go
func (z *ZFS) Unmount(req *acomm.Request) (interface{}, *url.URL, error)
```
Unmount mounts a zfs filesystem.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
