# simple

[![simple](https://godoc.org/github.com/mistifyio/provider-simple?status.png)](https://godoc.org/github.com/mistifyio/provider-simple)

Package simple is a simple task provider proof of concept.

## Usage

#### type CPUInfo

```go
type CPUInfo struct {
	Processor int `json:"processor"`
	MHz       int `json:"mhz"`
}
```

CPUInfo is information on a particular CPU.

#### type CPUInfoArgs

```go
type CPUInfoArgs struct {
	GuestID string `json:"guest_id"`
}
```

CPUInfoArgs are arguments for the CPUInfo handler.

#### type CPUInfoResult

```go
type CPUInfoResult []*CPUInfo
```

CPUInfoResult is the result data for the CPUInfo handler.

#### type Config

```go
type Config struct {
}
```

Config holds all configuration for the provider.

#### func  NewConfig

```go
func NewConfig(v *viper.Viper) *Config
```
NewConfig creates a new instance of Config. If a viper instance is not provided,
a new one will be created.

#### func (*Config) CoordinatorURL

```go
func (c *Config) CoordinatorURL() *url.URL
```
CoordinatorURL returns the URL of the Coordinator for which the Provider is
registered.

#### func (*Config) LoadConfigFile

```go
func (c *Config) LoadConfigFile() error
```
LoadConfigFile attempts to load a config file.

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

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type DiskInfo

```go
type DiskInfo struct {
	Device string
	Size   int64
}
```

DiskInfo is information on a particular disk.

#### type DiskInfoArgs

```go
type DiskInfoArgs struct {
	GuestID string `json:"guest_id"`
}
```

DiskInfoArgs are arguments for the DiskInfo handler.

#### type DiskInfoResult

```go
type DiskInfoResult []*DiskInfo
```

DiskInfoResult is the result data for the DiskInfo handler.

#### type MultiRequest

```go
type MultiRequest struct {
}
```

MultiRequest provides a way to manage multiple parallel requests

#### func  NewMultiRequest

```go
func NewMultiRequest(tracker *acomm.Tracker) *MultiRequest
```
NewMultiRequest creates and initializes a new MultiRequest.

#### func (*MultiRequest) AddRequest

```go
func (m *MultiRequest) AddRequest(name string, req *acomm.Request) error
```
AddRequest adds a request to the MultiRequest. Sending the request is still the
responsibility of the caller.

#### func (*MultiRequest) RemoveRequest

```go
func (m *MultiRequest) RemoveRequest(req *acomm.Request)
```
RemoveRequest removes a request from the MultiRequest. Useful if the send fails.

#### func (*MultiRequest) Responses

```go
func (m *MultiRequest) Responses() map[string]*acomm.Response
```
Responses returns responses for all of the requests, keyed on the request name
(as opposed to request id). Blocks until all requests are accounted for.

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

#### type Simple

```go
type Simple struct {
}
```

Simple is a simple provider implementation.

#### func  NewSimple

```go
func NewSimple(config *Config, tracker *acomm.Tracker) *Simple
```
NewSimple creates a new instance of Simple.

#### func (*Simple) CPUInfo

```go
func (s *Simple) CPUInfo(req *acomm.Request) (interface{}, *url.URL, error)
```
CPUInfo is a task handler to retrieve information about CPUs.

#### func (*Simple) DiskInfo

```go
func (s *Simple) DiskInfo(req *acomm.Request) (interface{}, *url.URL, error)
```
DiskInfo is a task handler to retrieve information about disks.

#### func (*Simple) RegisterTasks

```go
func (s *Simple) RegisterTasks(server *Server)
```
RegisterTasks registers all of Simple's task handlers with the server.

#### func (*Simple) StreamEcho

```go
func (s *Simple) StreamEcho(req *acomm.Request) (interface{}, *url.URL, error)
```
StreamEcho is a task handler to echo input back via streaming data

#### func (*Simple) SystemStatus

```go
func (s *Simple) SystemStatus(req *acomm.Request) (interface{}, *url.URL, error)
```
SystemStatus is a task handler to retrieve info look up and return system
information. It depends on and makes requests for several other tasks.

#### type SystemStatusArgs

```go
type SystemStatusArgs struct {
	GuestID string `json:"guest_id"`
}
```

SystemStatusArgs are arguments for the SystemStatus handler.

#### type SystemStatusResult

```go
type SystemStatusResult struct {
	CPUs  []*CPUInfo  `json:"cpus"`
	Disks []*DiskInfo `json:"disks"`
}
```

SystemStatusResult is the result data for the SystemStatus handler.

#### type TaskHandler

```go
type TaskHandler func(*acomm.Request) (interface{}, *url.URL, error)
```

TaskHandler if the request handler function for a particular task. It should
return results or an error, but not both.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
