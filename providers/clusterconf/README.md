# clusterconf

[![clusterconf](https://godoc.org/github.com/cerana/cerana/providers/clusterconf?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/clusterconf)



## Usage

```go
const (
	RWZFS = iota
	TempZFS
	RAMDisk
)
```
Valid bundle dataset types

#### type Bundle

```go
type Bundle struct {
	ID         uint64                   `json:"id"`
	Datasets   map[string]BundleDataset `json:"datasets"`
	Services   map[string]BundleService `json:"services"`
	Redundancy int                      `json:"redundancy"`
	Ports      BundlePorts              `json:"ports"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}
```

Bundle is information about a bundle of services.

#### type BundleDataset

```go
type BundleDataset struct {
	Name  string            `json:"name"`
	ID    string            `json:"id"`
	Type  BundleDatasetType `json:"type"`
	Quota int               `json:"type"`
}
```

BundleDataset is configuration for a dataset associated with a bundle.

#### type BundleDatasetType

```go
type BundleDatasetType int
```

BundleDatasetType is the type of dataset to be used in a bundle.

#### type BundleHeartbeat

```go
type BundleHeartbeat struct {
	IP           net.IP           `json:"ip"`
	HealthErrors map[string]error `json:"healthErrors"`
}
```


#### func (BundleHeartbeat) MarshalJSON

```go
func (b BundleHeartbeat) MarshalJSON() ([]byte, error)
```
MarshalJSON marshals BundleHeartbeat into a JSON map, converting error values to
strings.

#### func (*BundleHeartbeat) UnmarshalJSON

```go
func (b *BundleHeartbeat) UnmarshalJSON(data []byte) error
```
UnmarshalJSON unmarshals JSON into a BundleHeartbeat, converting string values
to errors.

#### type BundleHeartbeatArgs

```go
type BundleHeartbeatArgs struct {
	ID           uint64           `json:"id"`
	Serial       string           `json:"serial"`
	IP           net.IP           `json:"ip"`
	HealthErrors map[string]error `json:"healthErrors"`
}
```


#### type BundleHeartbeatList

```go
type BundleHeartbeatList struct {
	Heartbeats map[uint64]BundleHeartbeats `json:"heartbeats"`
}
```


#### func (BundleHeartbeatList) MarshalJSON

```go
func (b BundleHeartbeatList) MarshalJSON() ([]byte, error)
```
MarshalJSON marshals BundleHeartbeatList into a JSON map, converting uint keys
to strings.

#### func (*BundleHeartbeatList) UnmarshalJSON

```go
func (b *BundleHeartbeatList) UnmarshalJSON(data []byte) error
```
UnmarshalJSON unmarshals JSON into a BundleHeartbeatList, converting string keys
to uints.

#### type BundleHeartbeats

```go
type BundleHeartbeats map[string]BundleHeartbeat
```


#### type BundleListResult

```go
type BundleListResult struct {
	Bundles []*Bundle `json:"bundles"`
}
```

BundleListResult is the result from listing bundles.

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
type BundlePorts map[int]BundlePort
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
	ServiceConf
	Datasets map[string]ServiceDataset `json:"datasets"`
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
BundleHeartbeat registers a new node heartbeat that is using the dataset.

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

#### func (*ClusterConf) ListBundleHeartbeats

```go
func (c *ClusterConf) ListBundleHeartbeats(req *acomm.Request) (interface{}, *url.URL, error)
```

#### func (*ClusterConf) ListBundles

```go
func (c *ClusterConf) ListBundles(req *acomm.Request) (interface{}, *url.URL, error)
```
ListBundles retrieves a list of all bundles.

#### func (*ClusterConf) ListDatasetHeartbeats

```go
func (c *ClusterConf) ListDatasetHeartbeats(req *acomm.Request) (interface{}, *url.URL, error)
```

#### func (*ClusterConf) ListDatasets

```go
func (c *ClusterConf) ListDatasets(req *acomm.Request) (interface{}, *url.URL, error)
```
ListDatasets returns a list of all Datasets.

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

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the ClusterConf provider config.

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
	DatasetTTL string `json:"datasetTTL"`
	BundleTTL  string `json:"bundleTTL"`
	NodeTTL    string `json:"nodeTTL"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type Dataset

```go
type Dataset struct {
	ID                string `json:"id"`
	Parent            string `json:"parent"`
	ParentSameMachine bool   `json:"parentSameMachine"`
	ReadOnly          bool   `json:"readOnly"`
	NFS               bool   `json:"nfs"`
	Redundancy        int    `json:"redundancy"`
	Quota             int    `json:"quota"`
	// ModIndex should be treated as opaque, but passed back on updates.
	ModIndex uint64 `json:"modIndex"`
}
```

Dataset is information about a dataset.

#### type DatasetHeartbeat

```go
type DatasetHeartbeat struct {
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}
```


#### type DatasetHeartbeatArgs

```go
type DatasetHeartbeatArgs struct {
	ID    string `json:"id"`
	IP    net.IP `json:"ip"`
	InUse bool   `json:"inUse"`
}
```

DatasetHeartbeatArgs are arguments for updating a dataset node heartbeat.

#### type DatasetHeartbeatList

```go
type DatasetHeartbeatList struct {
	Heartbeats map[string]map[string]DatasetHeartbeat
}
```


#### type DatasetListResult

```go
type DatasetListResult struct {
	Datasets []*Dataset `json:"datasets"`
}
```

DatasetListResult is the result for listing datasets.

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
	DefaultsConf

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

#### type DeleteBundleArgs

```go
type DeleteBundleArgs struct {
	ID uint64 `json:"id"`
}
```

DeleteBundleArgs are args for bundle delete task.

#### type GetBundleArgs

```go
type GetBundleArgs struct {
	ID              uint64 `json:"id"`
	CombinedOverlay bool   `json:"overlay"`
}
```

GetBundleArgs are args for retrieving a bundle.

#### type HealthCheck

```go
type HealthCheck struct {
	ID   string      `json:"id"`
	Type string      `json:"type"`
	Args interface{} `json:"args"`
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

#### type ListBundleArgs

```go
type ListBundleArgs struct {
	CombinedOverlay bool `json:"overlay"`
}
```

ListBundleArgs are args for retrieving a bundle list.

#### type MockClusterConf

```go
type MockClusterConf struct {
	Data *MockClusterData
}
```

MockClusterConf is a mock ClusterConf provider.

#### func  NewMockClusterConf

```go
func NewMockClusterConf() *MockClusterConf
```
NewMockClusterConf creates a new MockClusterConf.

#### func (*MockClusterConf) BundleHeartbeat

```go
func (c *MockClusterConf) BundleHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
BundleHeartbeat adds a mock bundle heartbeat.

#### func (*MockClusterConf) DatasetHeartbeat

```go
func (c *MockClusterConf) DatasetHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
DatasetHeartbeat adds a mock dataset heartbeat.

#### func (*MockClusterConf) DeleteBundle

```go
func (c *MockClusterConf) DeleteBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteBundle removes a mock bundle.

#### func (*MockClusterConf) DeleteDataset

```go
func (c *MockClusterConf) DeleteDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteDataset removes a mock dataset.

#### func (*MockClusterConf) DeleteService

```go
func (c *MockClusterConf) DeleteService(req *acomm.Request) (interface{}, *url.URL, error)
```
DeleteService removes a mock service.

#### func (*MockClusterConf) GetBundle

```go
func (c *MockClusterConf) GetBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
GetBundle retrieves a mock bundle.

#### func (*MockClusterConf) GetDataset

```go
func (c *MockClusterConf) GetDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
GetDataset retrieves a mock dataset.

#### func (*MockClusterConf) GetDefaults

```go
func (c *MockClusterConf) GetDefaults(req *acomm.Request) (interface{}, *url.URL, error)
```
GetDefaults retrieves the mock default values.

#### func (*MockClusterConf) GetNode

```go
func (c *MockClusterConf) GetNode(req *acomm.Request) (interface{}, *url.URL, error)
```
GetNode retrieves a mock node.

#### func (*MockClusterConf) GetNodesHistory

```go
func (c *MockClusterConf) GetNodesHistory(req *acomm.Request) (interface{}, *url.URL, error)
```
GetNodesHistory retrieves mock nodes history.

#### func (*MockClusterConf) GetService

```go
func (c *MockClusterConf) GetService(req *acomm.Request) (interface{}, *url.URL, error)
```
GetService retrieves a mock service.

#### func (*MockClusterConf) ListBundles

```go
func (c *MockClusterConf) ListBundles(req *acomm.Request) (interface{}, *url.URL, error)
```
ListBundles retrieves all mock bundles.

#### func (*MockClusterConf) ListDatasets

```go
func (c *MockClusterConf) ListDatasets(req *acomm.Request) (interface{}, *url.URL, error)
```
ListDatasets lists all mock datasets.

#### func (*MockClusterConf) NodeHeartbeat

```go
func (c *MockClusterConf) NodeHeartbeat(req *acomm.Request) (interface{}, *url.URL, error)
```
NodeHeartbeat adds a mock node heartbeat.

#### func (*MockClusterConf) RegisterTasks

```go
func (c *MockClusterConf) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of MockClusterConf's tasks.

#### func (*MockClusterConf) UpdateBundle

```go
func (c *MockClusterConf) UpdateBundle(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateBundle updates a mock bundle.

#### func (*MockClusterConf) UpdateDataset

```go
func (c *MockClusterConf) UpdateDataset(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateDataset updates a mock dataset.

#### func (*MockClusterConf) UpdateDefaults

```go
func (c *MockClusterConf) UpdateDefaults(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateDefaults updates the mock default values.

#### func (*MockClusterConf) UpdateService

```go
func (c *MockClusterConf) UpdateService(req *acomm.Request) (interface{}, *url.URL, error)
```
UpdateService updates a mock service.

#### type MockClusterData

```go
type MockClusterData struct {
	Services  map[string]*Service
	Bundles   map[uint64]*Bundle
	BundlesHB map[uint64]BundleHeartbeats
	Datasets  map[string]*Dataset
	DatasetHB map[string]bool
	Nodes     map[string]*Node
	History   NodesHistory
	Defaults  *Defaults
}
```

MockClusterData is the in-memory data structure for a MockClusterConf.

#### type Node

```go
type Node struct {
	ID          string       `json:"id"`
	Heartbeat   time.Time    `json:"heartbeat"`
	MemoryTotal uint64       `json:"memoryTotal"`
	MemoryFree  uint64       `json:"memoryFree"`
	CPUCores    int          `json:"cpuCores"`
	CPULoad     load.AvgStat `json:"cpuLoad"`
	DiskTotal   uint64       `json:"diskTotal"`
	DiskFree    uint64       `json:"diskFree"`
}
```

Node is current information about a hardware node.

#### type NodeHistory

```go
type NodeHistory map[time.Time]*Node
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
	History NodesHistory `json:"history"`
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
	ServiceConf

	// ModIndex should be treated as opaque, but passed back on updates
	ModIndex uint64 `json:"modIndex"`
}
```

Service is information about a service.

#### type ServiceConf

```go
type ServiceConf struct {
	ID           string                 `json:"id"`
	Dataset      string                 `json:"dataset"`
	HealthChecks map[string]HealthCheck `json:"healthChecks"`
	Limits       ResourceLimits         `json:"limits"`
	Env          map[string]string      `json:"env"`
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
