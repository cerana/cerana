# kv

[![kv](https://godoc.org/github.com/cerana/cerana/providers/kv?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/kv)



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

#### func (*Config) Address

```go
func (c *Config) Address() (string, error)
```
Address returns the configured address of the consul kv server.

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the KV provider config.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type Cookie

```go
type Cookie struct {
	Cookie uint64 `json:"cookie"`
}
```

Cookie holds a unique value that is used as a reference to server side storage.

#### type DeleteArgs

```go
type DeleteArgs struct {
	Key       string `json:"key"`
	Recursive bool   `json:"recursive"`
}
```

DeleteArgs specify the arguments to the "kv-delete" endpoint.

#### type EphemeralDestroyArgs

```go
type EphemeralDestroyArgs struct {
	Key string `json:"key"`
}
```

EphemeralDestroyArgs specifies the arguments to the "kv-ephemeral-destroy"
endpoint.

#### type EphemeralSetArgs

```go
type EphemeralSetArgs struct {
	Key   string        `json:"key"`
	Value string        `json:"value"`
	TTL   time.Duration `json:"ttl"`
}
```

EphemeralSetArgs specifies the arguments to the "kv-ephemeral-set" endpoint.

#### type Event

```go
type Event struct {
	kv.Event
	Error error
}
```

Event specifies structure describing events that took place on watched prefixes.

#### type GetArgs

```go
type GetArgs struct {
	Key string `json:"key"`
}
```

GetArgs specify the arguments to the "kv-get" endpoint.

#### type KV

```go
type KV struct {
}
```

KV is a provider of kv functionality.

#### func  New

```go
func New(config *Config, tracker *acomm.Tracker) (*KV, error)
```
New creates a new instance of KV.

#### func (*KV) RegisterTasks

```go
func (k *KV) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of KV's task handlers with the server.

#### type LockArgs

```go
type LockArgs struct {
	Key string        `json:"key"`
	TTL time.Duration `json:"ttl"`
}
```

LockArgs specifies the arguments to the "kv-lock" endpoint.

#### type Mock

```go
type Mock struct {
	*KV
}
```

Mock is a mock KV provider

#### func  NewMock

```go
func NewMock(config *provider.Config, tracker *acomm.Tracker) (*Mock, error)
```
NewMock starts up a kv backend server and instantiates a new kv.KV provider. The
kv backend is started on the port provided as part of config.Address().
Mock.Stop() should be called when testing is done in order to clean up.

#### func (*Mock) Stop

```go
func (m *Mock) Stop()
```
Stop will stop the kv and remove the temporary directory used for it's data

#### type RemoveArgs

```go
type RemoveArgs struct {
	Key   string `json:"key"`
	Index uint64 `json:"index"`
}
```

RemoveArgs specifies the arguments to the "kv-remove" endpoint.

#### type SetArgs

```go
type SetArgs struct {
	Key  string `json:"key"`
	Data string `json:"string"`
}
```

SetArgs specify the arguments to the "kv-set" endpoint.

#### type UpdateArgs

```go
type UpdateArgs struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Index uint64 `json:"index"`
}
```

UpdateArgs specifies the arguments to the "kv-update" endpoint.

#### type UpdateReturn

```go
type UpdateReturn struct {
	Index uint64 `json:"index"`
}
```

UpdateReturn specifies the return value from the "kv-update" endpoint.

#### type WatchArgs

```go
type WatchArgs struct {
	Prefix string `json:"prefix"`
	Index  uint64 `json:"index"`
}
```

WatchArgs specify the arguments to the "kv-watch" endpoint.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
