# service

[![service](https://godoc.org/github.com/cerana/cerana/providers/service?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/service)



## Usage

#### type Config

```go
type Config struct {
	*provider.Config
}
```

Config holds all configuration for the provider.

#### func  NewConfig

```go
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config
```
NewConfig creates a new instance of Config.

#### func (*Config) DatasetCloneDir

```go
func (c *Config) DatasetCloneDir() string
```
DatasetCloneDir returns the zfs path in which to clone datasets.

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the config data.

#### func (*Config) RollbackCloneCmd

```go
func (c *Config) RollbackCloneCmd() string
```
RollbackCloneCmd returns the full path of the clone/rollback script datasets for
services.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type ConfigData

```go
type ConfigData struct {
	provider.ConfigData
	RollbackCloneCmd string `json:"rollback_clone_cmd"`
	DatasetCloneDir  string `json:"dataset_clone_dir"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type CreateArgs

```go
type CreateArgs struct {
	ID          string            `json:"id"`
	BundleID    uint64            `json:"bundleID"`
	Dataset     string            `json:"dataset"`
	Description string            `json:"description"`
	Cmd         []string          `json:"cmd"`
	Env         map[string]string `json:"env"`
	Overwrite   bool              `json:"overwrite"`
}
```

CreateArgs contains args for creating or replacing a Service.

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

#### type Mock

```go
type Mock struct {
	Data MockData
}
```

Mock is a mock provider of service management functionality.

#### func  NewMock

```go
func NewMock() *Mock
```
NewMock creates a new instance of Mock.

#### func (*Mock) Add

```go
func (m *Mock) Add(service Service)
```
Add is a convenience method to directly add a mock service.

#### func (*Mock) ClearData

```go
func (m *Mock) ClearData()
```
ClearData clears all mock data.

#### func (*Mock) Create

```go
func (m *Mock) Create(req *acomm.Request) (interface{}, *url.URL, error)
```
Create creates a new mock service.

#### func (*Mock) Get

```go
func (m *Mock) Get(req *acomm.Request) (interface{}, *url.URL, error)
```
Get retrieves a mock service.

#### func (*Mock) List

```go
func (m *Mock) List(req *acomm.Request) (interface{}, *url.URL, error)
```
List lists all mock services.

#### func (*Mock) RegisterTasks

```go
func (m *Mock) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of the mock provider task handlers with the server.

#### func (*Mock) Remove

```go
func (m *Mock) Remove(req *acomm.Request) (interface{}, *url.URL, error)
```
Remove removes a mock service.

#### func (*Mock) Restart

```go
func (m *Mock) Restart(req *acomm.Request) (interface{}, *url.URL, error)
```
Restart restarts a mock service.

#### type MockData

```go
type MockData struct {
	Services map[uint64]map[string]Service
}
```

MockData is the in-memory data structure for the Mock.

#### type Provider

```go
type Provider struct {
}
```

Provider is a provider of service management functionality.

#### func  New

```go
func New(config *Config, tracker *acomm.Tracker) *Provider
```
New creates a new instance of Provider.

#### func (*Provider) Create

```go
func (p *Provider) Create(req *acomm.Request) (interface{}, *url.URL, error)
```
Create creates (or replaces) and starts (or restarts) a service.

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
	Cmd         []string          `json:"cmd"`
	UID         uint64            `json:"uid"`
	GID         uint64            `json:"gid"`
	Env         map[string]string `json:"env"`
}
```

Service is information about a service.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
