# clusterconf

[![clusterconf](https://godoc.org/github.com/cerana/cerana/providers/clusterconf?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/clusterconf)



## Usage

#### type Bundle

```go
type Bundle struct {
	*BundleConf

	// Nodes contains the set of nodes on which the dataset is currently in use.
	// THe map keys are serials.
	Nodes map[string]net.IP `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}
```

Bundle is information about a bundle of services.

#### type BundleConf

```go
type BundleConf struct {
	ID         int                       `json:"id"`
	Datasets   map[string]*BundleDataset `json:"datasets"`
	Services   map[string]*BundleService `json:"services"`
	Redundancy int                       `json:"redundancy"`
	Ports      BundlePorts               `json:"ports"`
}
```

BundleConf is the configuration of a bundle.

#### type BundleDataset

```go
type BundleDataset struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Type  int    `json:"type"` // TODO: Decide on type for this. Iota?
	Quota int    `json:"type"`
}
```

BundleDataset is configuration for a dataset associated with a bundle.

#### type BundleHeartbeatArgs

```go
type BundleHeartbeatArgs struct {
	ID     int    `json:"id"`
	Serial string `json:"serial"`
	IP     net.IP `json:"ip"`
}
```

BundleHeartbeatArgs are arguments for updating a dataset node heartbeat.

#### type BundleIDArgs

```go
type BundleIDArgs struct {
	ID int `json:"id"`
}
```

BundleIDArgs are args for bundle tasks that only require bundle id.

#### type BundlePayload

```go
type BundlePayload struct {
	Bundle *Bundle `json:"bundle"`
}
```

BundlePayload can be used for task args or result when a bundle object needs to
be sent.

#### type BundlePort

```go
type BundlePort struct {
	Port             int      `json:"port"`
	Public           bool     `json:"public"`
	ConnectedBundles []string `json:"connectedBundles"`
	ExternalPort     int      `json:"externalPort"`
}
```

BundlePort is configuration for a port associated with a bundle.

#### type BundlePorts

```go
type BundlePorts map[int]*BundlePort
```

BundlePorts is a map of port numbers to port information.

#### func (BundlePorts) MarshalJSON

```go
func (p BundlePorts) MarshalJSON() ([]byte, error)
```
MarshalJSON marshals BundlePorts into a JSON map, converting int keys to
strings.

#### func (BundlePorts) UnmarshalJSON

```go
func (p BundlePorts) UnmarshalJSON(data []byte) error
```
UnmarshalJSON unmarshals JSON into a BundlePorts, converting string keys to
ints.

#### type BundleService

```go
type BundleService struct {
	*ServiceConf
	Datasets map[string]*ServiceDataset `json:"datasets"`
}
```

BundleService is configuration overrides for a service of a bundle and
associated bundles.

#### type ClusterConf

```go
type ClusterConf struct {
}
```

ClusterConf is a provider of cluster configuration functionality.

#### func  New

```go
func New(config *Config, tracker *acomm.Tracker) *ClusterConf
```
New creates a new instance of ClusterConf

#### func (*ClusterConf) BundleHeartbeat

```go
func (c *ClusterConf) BundleHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
BundleHeartbeat registers a new node heartbeat that is using the bundle.

#### func (*ClusterConf) DatasetHeartbeat

```go
func (c *ClusterConf) DatasetHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
DatasetHeartbeat registers a new node heartbeat that is using the dataset.

#### func (*ClusterConf) DeleteBundle

```go
func (c *ClusterConf) DeleteBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteBundle deletes a bundle config.

#### func (*ClusterConf) DeleteDataset

```go
func (c *ClusterConf) DeleteDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteDataset deletes a dataset config.

#### func (*ClusterConf) DeleteService

```go
func (c *ClusterConf) DeleteService(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteService deletes a service config.

#### func (*ClusterConf) GetBundle

```go
func (c *ClusterConf) GetBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
GetBundle retrieves a bundle.

#### func (*ClusterConf) GetDataset

```go
func (c *ClusterConf) GetDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
GetDataset retrieves a dataset.

#### func (*ClusterConf) GetDefaults

```go
func (c *ClusterConf) GetDefaults(req *acomm.Request) (interface{}, *url.URL, error)
```
GetDefaults retrieves the cluster config.

#### func (*ClusterConf) GetNode

```go
func (c *ClusterConf) GetNode(req *acomm.Request) (interface{}, *url.URL, error)
```
GetNode returns the latest information about a node.

#### func (*ClusterConf) GetNodesHistory

```go
func (c *ClusterConf) GetNodesHistory(req *acomm.Request) (interface{}, *url.URL, error)
```
GetNodesHistory gets the heartbeat history for one or more nodes.

#### func (*ClusterConf) GetService

```go
func (c *ClusterConf) GetService(req *acomm.Request) (interface{}, *url.URL, error)
```
GetService retrieves a service.

#### func (*ClusterConf) NodeHeartbeat

```go
func (c *ClusterConf) NodeHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
NodeHeartbeat records a new node heartbeat.

#### func (*ClusterConf) RegisterTasks

```go
func (c *ClusterConf) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Systemd's task handlers with the server.

#### func (*ClusterConf) UpdateBundle

```go
func (c *ClusterConf) UpdateBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateBundle creates or updates a bundle config. When updating, a Get should
first be performed and the modified Bundle passed back.

#### func (*ClusterConf) UpdateDataset

```go
func (c *ClusterConf) UpdateDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateDataset creates or updates a dataset config. When updating, a Get should
first be performed and the modified Dataset passed back.

#### func (*ClusterConf) UpdateDefaults

```go
func (c *ClusterConf) UpdateDefaults(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateDefaults sets or updates

#### func (*ClusterConf) UpdateService

```go
func (c *ClusterConf) UpdateService(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateService creates or updates a service config. When updating, a Get should
first be performed and the modified Service passed back.

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

#### func (*Config) BundleTTL

```go
func (c *Config) BundleTTL() time.Duration
```
BundleTTL returns the TTL for bundle node heartbeats.

#### func (*Config) DatasetTTL

```go
func (c *Config) DatasetTTL() time.Duration
```
DatasetTTL returns the TTL for dataset node heartbeats.

#### func (*Config) NodeTTL

```go
func (c *Config) NodeTTL() time.Duration
```
NodeTTL returns the TTL for node heartbeats.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type ConfigData

```go
type ConfigData struct {
	provider.ConfigData
	DatasetTTL time.Duration `json:"datasetTTL"`
	BundleTTL  time.Duration `json:"bundleTTL"`
	NodeTTL    time.Duration `json:"NodeTTL"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type Dataset

```go
type Dataset struct {
	*DatasetConf

	// Nodes contains the set of nodes on which the dataset is currently in use.
	// The map keys are IP address strings.
	Nodes map[string]bool `json:"nodes"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}
```

Dataset is information about a dataset.

#### type DatasetConf

```go
type DatasetConf struct {
	ID                string `json:"id"`
	Parent            string `json:"parent"`
	ParentSameMachine bool   `json:"parentSameMachine"`
	ReadOnly          bool   `json:"readOnly"`
	NFS               bool   `json:"nfs"`
	Redundancy        int    `json:"redundancy"`
	Quota             int    `json:"quota"`
}
```

DatasetConf is the configuration of a dataset.

#### type DatasetHeartbeatArgs

```go
type DatasetHeartbeatArgs struct {
	ID string `json:"id"`
	IP net.IP `json:"ip"`
}
```

DatasetHeartbeatArgs are arguments for updating a dataset node heartbeat.

#### type DatasetPayload

```go
type DatasetPayload struct {
	Dataset *Dataset `json:"dataset"`
}
```

DatasetPayload can be used for task args or result when a dataset object needs
to be sent.

#### type Defaults

```go
type Defaults struct {
	*DefaultsConf

	ModIndex uint64 `json:"modIndex"`
}
```

Defaults is information about the cluster configuration.

#### type DefaultsConf

```go
type DefaultsConf struct {
	ZFSManual bool `json:"zfsManual"`
}
```

DefaultsConf is the configuration for the cluster.

#### type DefaultsPayload

```go
type DefaultsPayload struct {
	Defaults *Defaults `json:"defaults"`
}
```

DefaultsPayload can be used for task args or result when a cluster object needs
to be sent.

#### type HealthCheck

```go
type HealthCheck struct {
	ID               string   `json:"id"`
	ProtocolProvider string   `json:"protocolProvider"`
	Parameters       []string `json:"parameters"`
}
```

HealthCheck is configuration for performing a health check.

#### type IDArgs

```go
type IDArgs struct {
	ID string `json:"id"`
}
```

IDArgs are arguments for operations requiring only an ID.

#### type Node

```go
type Node struct {
	ID          string    `json:"id"`
	Heartbeat   time.Time `json:"heartbeat"`
	MemoryTotal int64     `json:"memoryTotal"`
	MemoryFree  int64     `json:"memoryFree"`
	CPUTotal    int       `json:"cpuTotal"`
	CPUFree     int       `json:"cpuFree"`
	DiskTotal   int       `json:"diskTotal"`
	DiskFree    int       `json:"diskFree"`
}
```

Node is current information about a hardware node.

#### type NodeHistory

```go
type NodeHistory map[time.Time]Node
```

NodeHistory is a set of historical information for a node.

#### type NodeHistoryArgs

```go
type NodeHistoryArgs struct {
	IDs    []string  `json:"ids"`
	Before time.Time `json:"before"`
	After  time.Time `json:"after"`
}
```

NodeHistoryArgs are arguments for filtering the historical results for nodes.

#### type NodePayload

```go
type NodePayload struct {
	Node *Node `json:"node"`
}
```

NodePayload can be used for task args or result when a node object needs to be
sent.

#### type NodesHistory

```go
type NodesHistory map[string]NodeHistory
```

NodesHistory is the historical information for multiple nodes.

#### type NodesHistoryResult

```go
type NodesHistoryResult struct {
	History *NodesHistory `json:"history"`
}
```

NodesHistoryResult is the result from the GetNodesHistory handler.

#### type ResourceLimits

```go
type ResourceLimits struct {
	CPU       int   `json:"cpu"`
	Memory    int64 `json:"memory"`
	Processes int   `json:"processes"`
}
```

ResourceLimits is configuration for resource upper bounds.

#### type Service

```go
type Service struct {
	*ServiceConf

	// ModIndex should be treated as opaque, but passed back on updates
	ModIndex uint64 `json:"modIndex"`
}
```

Service is information about a service.

#### type ServiceConf

```go
type ServiceConf struct {
	ID           string                  `json:"id"`
	Dataset      string                  `json:"dataset"`
	HealthChecks map[string]*HealthCheck `json:"healthCheck"`
	Limits       *ResourceLimits         `json:"limits"`
	Env          map[string]string       `json:"env"`
}
```

ServiceConf is the configuration of a service.

#### type ServiceDataset

```go
type ServiceDataset struct {
	Name       string `json:"name"`
	MountPoint string `json:"mountPoint"`
	ReadOnly   bool   `json:"readOnly"`
}
```

ServiceDataset is configuration for mounting a dataset for a bundle service.

#### type ServicePayload

```go
type ServicePayload struct {
	Service *Service `json:"service"`
}
```

ServicePayload can be used for task args or result when a service object needs
to be sent.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
