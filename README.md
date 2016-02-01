# provider

[![provider](https://godoc.org/github.com/mistifyio/provider?status.png)](https://godoc.org/github.com/mistifyio/provider)



## Usage

#### type Config

```go
type Config struct {
}
```

Config holds all configuration for the provider.

#### func  NewConfig

```go
func NewConfig(flagSet *flag.FlagSet, v *viper.Viper) *Config
```
NewConfig creates a new instance of Config. If a viper instance is not provided,
a new one will be created.

#### func (*Config) CoordinatorURL

```go
func (c *Config) CoordinatorURL() *url.URL
```
CoordinatorURL returns the URL of the Coordinator for which the Provider is
registered.

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig attempts to load the config. Flags should be parsed first.

#### func (*Config) RequestTimeout

```go
func (c *Config) RequestTimeout() time.Duration
```
RequestTimeout returns the duration of the default request timeout.

#### func (*Config) ServiceName

```go
func (c *Config) ServiceName() string
```
ServiceName returns the name the service should register as.

#### func (*Config) SetupLogging

```go
func (c *Config) SetupLogging() error
```
SetupLogging sets the log level and formatting.

#### func (*Config) SocketDir

```go
func (c *Config) SocketDir() string
```
SocketDir returns the base directory for task sockets.

#### func (*Config) StreamDir

```go
func (c *Config) StreamDir() string
```
StreamDir returns the directory for ad-hoc data stream sockets.

#### func (*Config) TaskPriority

```go
func (c *Config) TaskPriority(taskName string) int
```
TaskPriority determines the registration priority of a task. If a priority was
not explicitly configured for the task, it will return the default.

#### func (*Config) TaskTimeout

```go
func (c *Config) TaskTimeout(taskName string) time.Duration
```
TaskTimeout determines the timeout for a task. If a timeout was not explicitly
configured for the task, it will return the default.

#### func (*Config) Unmarshal

```go
func (c *Config) Unmarshal(rawVal interface{}) error
```
Unmarshal unmarshals the config into a struct.

#### func (*Config) UnmarshalKey

```go
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error
```
UnmarshalKey unmarshals a single config key into a struct.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type Provider

```go
type Provider interface {
	RegisterTasks(Server)
}
```

Provider is an interface to allow a provider to register its tasks with a
Server.

#### type Server

```go
type Server struct {
}
```

Server is the main server struct.

#### func  NewServer

```go
func NewServer(config *Config) (*Server, error)
```
NewServer creates and initializes a new Server.

#### func (*Server) RegisterTask

```go
func (s *Server) RegisterTask(taskName string, handler TaskHandler)
```
RegisterTask registers a new task and its handler with the server.

#### func (*Server) RegisteredTasks

```go
func (s *Server) RegisteredTasks() []string
```
RegisteredTasks returns a list of registered task names.

#### func (*Server) Start

```go
func (s *Server) Start() error
```
Start starts up all of the registered tasks and response handling

#### func (*Server) Stop

```go
func (s *Server) Stop()
```
Stop stops all of the registered tasks and response handling. Blocks until
complete.

#### func (*Server) StopOnSignal

```go
func (s *Server) StopOnSignal(signals ...os.Signal)
```
StopOnSignal will wait until one of the specified signals is received and then
stop the server. If no signals are specified, it will use a default set.

#### func (*Server) Tracker

```go
func (s *Server) Tracker() *acomm.Tracker
```
Tracker returns the request/response tracker of the Server.

#### type TaskHandler

```go
type TaskHandler func(*acomm.Request) (interface{}, *url.URL, error)
```

TaskHandler if the request handler function for a particular task. It should
return results or an error, but not both.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
