# datatrade

[![datatrade](https://godoc.org/github.com/cerana/cerana/providers/datatrade?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/datatrade)



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

#### func (*Config) DatasetDir

```go
func (c *Config) DatasetDir() string
```
DatasetDir returns the directory in which datasets are stored on nodes.

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the config data.

#### func (*Config) NodeCoordinatorPort

```go
func (c *Config) NodeCoordinatorPort() uint
```
NodeCoordinatorPort returns the port that node coordinators are running on.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type ConfigData

```go
type ConfigData struct {
	provider.ConfigData
	DatasetDir          string `json:"dataset_dir"`
	NodeCoordinatorPort uint   `json:"dataset_dir"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type DatasetImportArgs

```go
type DatasetImportArgs struct {
	NFS        bool `json:"nfs"`
	Quota      int  `json:"quota"`
	ReadOnly   bool `json:"readOnly"`
	Redundancy int  `json:"redundancy"`
}
```

DatasetImportArgs are arguments for configuring an imported dataset.

#### type DatasetImportResult

```go
type DatasetImportResult struct {
	Dataset clusterconf.Dataset `json:"dataset"`
	NodeID  string              `json:"nodeID"`
}
```

DatasetImportResult is the result of a dataset import.

#### type Provider

```go
type Provider struct {
}
```

Provider is a provider of data import and export functionality.

#### func  New

```go
func New(config *Config, tracker *acomm.Tracker) *Provider
```
New creates a new instance of Provider.

#### func (*Provider) DatasetImport

```go
func (p *Provider) DatasetImport(req *acomm.Request) (interface{}, *url.URL, error)
```
DatasetImport imports a dataset into the cluster and tracks it in the cluster
configuration.

#### func (*Provider) RegisterTasks

```go
func (p *Provider) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of the provider task handlers with the server.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
