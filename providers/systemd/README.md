# systemd

[![systemd](https://godoc.org/github.com/cerana/cerana/providers/systemd?status.png)](https://godoc.org/github.com/cerana/cerana/providers/systemd)



## Usage

```go
const (
	ModeReplace    = "replace"
	ModeFail       = "fail"
	ModeIsolate    = "isolate"
	ModeIgnoreDeps = "ignore-dependencies"
	ModeIgnoreReqs = "ignore-requirements"
)
```
Unit start modes.

#### type ActionArgs

```go
type ActionArgs struct {
	Name string `json:"name"`
	Mode string `json:"mode"`
}
```

ActionArgs are arguments for service running action handlers.

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

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the Systemd provider config

#### func (*Config) UnitFilePath

```go
func (c *Config) UnitFilePath(name string) (string, error)
```
UnitFilePath returns the absolute path of the unit file for a unit.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type ConfigData

```go
type ConfigData struct {
	provider.ConfigData
	UnitFileDir string `json:"unit_file_dir"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type CreateArgs

```go
type CreateArgs struct {
	Name        string             `json:"name"`
	UnitOptions []*unit.UnitOption `json:"unit-options"`
}
```

CreateArgs are arguments for the Create handler.

#### type DisableArgs

```go
type DisableArgs struct {
	Name    string `json:"name"`
	Runtime bool   `json:"runtime"`
}
```

DisableArgs are arguments for the disable handler.

#### type EnableArgs

```go
type EnableArgs struct {
	Name    string `json:"name"`
	Runtime bool   `json:"runtime"`
	Force   bool   `json:"force"`
}
```

EnableArgs are arguments for the disable handler.

#### type GetArgs

```go
type GetArgs struct {
	Name string `json:"name"`
}
```

GetArgs are args for the Get handler

#### type GetResult

```go
type GetResult struct {
	Unit dbus.UnitStatus `json:"unit"`
}
```

GetResult is the result of the ListUnits handler.

#### type ListResult

```go
type ListResult struct {
	Units []dbus.UnitStatus `json:"units"`
}
```

ListResult is the result of the List handler.

#### type RemoveArgs

```go
type RemoveArgs struct {
	Name string `json:"name"`
}
```

RemoveArgs are arguments for the Remove handler.

#### type Systemd

```go
type Systemd struct {
}
```

Systemd is a provider of systemd functionality.

#### func  New

```go
func New(config *Config) (*Systemd, error)
```
New creates a new instance of Systemd.

#### func (*Systemd) Create

```go
func (s *Systemd) Create(req *acomm.Request) (interface{}, *url.URL, error)
```
Create creates a new unit file.

#### func (*Systemd) Disable

```go
func (s *Systemd) Disable(req *acomm.Request) (interface{}, *url.URL, error)
```
Disable disables a service.

#### func (*Systemd) Enable

```go
func (s *Systemd) Enable(req *acomm.Request) (interface{}, *url.URL, error)
```
Enable disables a service.

#### func (*Systemd) Get

```go
func (s *Systemd) Get(req *acomm.Request) (interface{}, *url.URL, error)
```
Get retuns a list of unit statuses.

#### func (*Systemd) List

```go
func (s *Systemd) List(req *acomm.Request) (interface{}, *url.URL, error)
```
List retuns a list of unit statuses.

#### func (*Systemd) RegisterTasks

```go
func (s *Systemd) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Systemd's task handlers with the server.

#### func (*Systemd) Remove

```go
func (s *Systemd) Remove(req *acomm.Request) (interface{}, *url.URL, error)
```
Remove deletes a unit file.

#### func (*Systemd) Restart

```go
func (s *Systemd) Restart(req *acomm.Request) (interface{}, *url.URL, error)
```
Restart restarts a service.

#### func (*Systemd) Start

```go
func (s *Systemd) Start(req *acomm.Request) (interface{}, *url.URL, error)
```
Start starts an enabled service.

#### func (*Systemd) Stop

```go
func (s *Systemd) Stop(req *acomm.Request) (interface{}, *url.URL, error)
```
Stop stops a running service.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
