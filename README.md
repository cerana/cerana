# provider

[![provider](https://godoc.org/github.com/mistifyio/provider?status.png)](https://godoc.org/github.com/mistifyio/provider)

Package provider is the framework for creating Providers. A Provider offers a
set of related functionality in the form of tasks, which Coordinators use
individually or in combination, to accomplish actions.


### Functionality Registration

Coordinators need to know what providers are available and what they offer in
order to complete tasks. Instead of an active registration, which requires
heartbeats and deregistration, Providers utilize the filesystem. For
registration of each task, a unix socket is created in a known directory
structure based on the task name. A Coordinator can check a task directory for
the socket when routing requests. When a Provider is shut down, the unix socket
will be automatically removed, effectively de-registering from the Coordinator.
In order to handle multiple Providers capable of handling the same task, socket
filenames are prefixed with a priority value.

Task socket path: `/[socket_dir]/[task_name]/[priority]-[provider-name].sock`


### Communication

Communication is handled via the acomm package. An initial request is received
on a task-specific socket. A response is sent to acknowledge receipt on the same
connection before the request is sent to the appropriate TaskHandler. After
completing the work, the TaskHandler returns a combination of result, data
stream url, and error. These three are bundled into a response and sent to the
request's responseHook. In the case of data streaming, the caller will connect
to the stream url and stream the data.

All requests originating from a provider go through the coordinator; providers
should not make requests directly to each other. These requests should be
tracked with the tracker. Responses may be sent directly to a unix socket
response hook. Data streaming is also handled directly. While providers could
respond to external (http) response hooks and stream data directly externally,
they will be separated by the coordinator, through which responses will be
proxied. Providers are free, however, to make requests to external services.


### Creating A Provider

As the description states, the provider package is not itself a complete
provider; it is used to create one. The process is quite simple. A provider
implementation will have a set of functions following the TaskHandler signature,
which accepts a request and returns components for a response. If any tasks are
going to make additional requests, the provider server's tracker should be used.
The provider must follow the Provider interface, which requires a method to
register the task handlers; the method should accept a Server and call its
RegisterTask method for each task handler.

Create a new Config, optionally supplying a flagset or viper instance. If
flagset is not provided, it will use the commandline flagset, and if viper is
not provided, one will be created. Flags should then be parsed, and the Config
loaded. After loading the Config, create a new Server using it. Then
create/initialize the provider, and register the tasks. Finally, start the
server and either use StopOnSignal or another way to wait and Stop the server
appropriately.

An example provider can be found in the examples directory.


### Config

There are a number of values required in the config for a provider to operate
successfully. The Config struct will add a number of the config options as flags
(including `config_file`). Tasks without explicit config for priority will use
default value.

    {
    	"config_file": "/path/to/config/file.json",
    	"socket_dir": "/base/path/for/sockets",
    	"default_priority": 50,
    	"log_level": "warning",
    	"request_timeout": 0,
    	"tasks":{
    		"ATaskNameFoo":{
    			"priority": 60,
    		}
    	}
    }


### Suggestions

Task handlers should be kept focused and self-contained as possible, doing one
logical operation. Use additional task requests to compose these focused
building blocks into actions.

All non-informational tasks should have a corrosponding reverse or cleanup task.
Make use of request error handlers to call such cleanup tasks.

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
func (c *Config) StreamDir(taskName string) string
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

#### type ConfigData

```go
type ConfigData struct {
	SocketDir       string                     `json:"socket_dir"`
	ServiceName     string                     `json:"service_name"`
	CoordinatorURL  string                     `json:"coordinator_url"`
	DefaultPriority uint                       `json:"default_priority"`
	LogLevel        string                     `json:"log_level"`
	DefaultTimeout  uint64                     `json:"default_timeout"`
	RequestTimeout  uint64                     `json:"request_timeout"`
	Tasks           map[string]*TaskConfigData `json:"tasks"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

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

#### type TaskConfigData

```go
type TaskConfigData struct {
	Priority uint   `json:"priority"`
	Timeout  uint64 `json:"timeout"`
}
```

TaskConfigData defines the structure of the task config data (e.g. in the config
file)

#### type TaskHandler

```go
type TaskHandler func(*acomm.Request) (interface{}, *url.URL, error)
```

TaskHandler if the request handler function for a particular task. It should
return results or an error, but not both.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
