# namespace

[![namespace](https://godoc.org/github.com/cerana/cerana/providers/namespace?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/namespace)



## Usage

#### type IDMap

```go
type IDMap struct {
	ID     uint64 `json:"id"`
	HostID uint64 `json:"hostID"`
	Length uint64 `json:"length"`
}
```

IDMap is a map of id in container to id on host and length of a range.

#### type Mock

```go
type Mock struct {
	Data MockData
}
```

Mock is a mock Namespace provider.

#### func  NewMock

```go
func NewMock() *Mock
```
NewMock creates a new instance of Mock.

#### func (*Mock) RegisterTasks

```go
func (n *Mock) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Mock's task handlers.

#### func (*Mock) SetUser

```go
func (n *Mock) SetUser(req *acomm.Request) (interface{}, *url.URL, error)
```
SetUser sets mock uid and gid mappings.

#### type MockData

```go
type MockData struct {
	SetUserErr error
}
```

MockData is the in-memory data structure for a Mock.

#### type Namespace

```go
type Namespace struct {
}
```

Namespace is a provider of namespace functionality.

#### func  New

```go
func New(config *provider.Config, tracker *acomm.Tracker) *Namespace
```
New creates a new instance of Namespace.

#### func (*Namespace) RegisterTasks

```go
func (n *Namespace) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Namespaces's task handlers with the server.

#### func (*Namespace) SetUser

```go
func (n *Namespace) SetUser(req *acomm.Request) (interface{}, *url.URL, error)
```
SetUser sets the user and group id mapping for a process.

#### type UserArgs

```go
type UserArgs struct {
	PID  uint64  `json:"pid"`
	UIDs []IDMap `json:"uids"`
	GIDs []IDMap `json:"gids"`
}
```

UserArgs are arguments for SetUser.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
