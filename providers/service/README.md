# service

[![service](https://godoc.org/github.com/cerana/cerana/providers/service?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/service)



## Usage

#### type CreateArgs

```go
type CreateArgs struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Description string            `json:"description"`
	Exec        []string          `json:"exec"`
	Env         map[string]string `json:"env"`
}
```

CreateArgs contains args for creating a new Service.

#### type GetArgs

```go
type GetArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}
```

GetArgs are args for retrieving a service.

#### type GetResult

```go
type GetResult struct {
	Service Service `json:"service"`
}
```

GetResult is the result of a Get.

#### type ListResult

```go
type ListResult struct {
	Services []Service
}
```

ListResult is the result of the List handler.

#### type Provider

```go
type Provider struct {
}
```

Provider is a provider of service management functionality.

#### func  New

```go
func New(config *provider.Config, tracker *acomm.Tracker) *Provider
```
New creates a new instance of Provider.

#### func (*Provider) Create

```go
func (p *Provider) Create(req *acomm.Request) (interface{}, *url.URL, error)
```
Create creates and starts a new service.

#### func (*Provider) Get

```go
func (p *Provider) Get(req *acomm.Request) (interface{}, *url.URL, error)
```
Get retrieves a service.

#### func (*Provider) List

```go
func (p *Provider) List(req *acomm.Request) (interface{}, *url.URL, error)
```
List returns a list of Services and information about each.

#### func (*Provider) RegisterTasks

```go
func (p *Provider) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of the provider task handlers with the server.

#### func (*Provider) Remove

```go
func (p *Provider) Remove(req *acomm.Request) (interface{}, *url.URL, error)
```
Remove removes a service from the node.

#### func (*Provider) Restart

```go
func (p *Provider) Restart(req *acomm.Request) (interface{}, *url.URL, error)
```
Restart restarts a service.

#### type RemoveArgs

```go
type RemoveArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}
```

RemoveArgs are arguments for the Remove task.

#### type RestartArgs

```go
type RestartArgs struct {
	ID       string `json:"id"`
	BundleID uint64 `json:"bundleID"`
}
```

RestartArgs are arguments for Restart.

#### type Service

```go
type Service struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Description string            `json:"description"`
	Uptime      time.Duration     `json:"uptime"`
	ActiveState string            `json:"activeState"`
	Exec        []string          `json:"exec"`
	UID         uint64            `json:"uid"`
	GID         uint64            `json:"gid"`
	Env         map[string]string `json:"env"`
}
```

Service is information about a service.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
