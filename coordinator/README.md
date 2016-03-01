# coordinator

[![coordinator](https://godoc.org/github.com/mistifyio/mistify/coordinator?status.png)](https://godoc.org/github.com/mistifyio/mistify/coordinator)

Package coordinator is a request router and proxy, to be used with Providers. It
is the entry point for all task requests and separates local providers from
external services.

New task requests are first sent to a Coordinator. There are two means the
Coordinator has of receiving requests: an http server, used for external
requests, and a unix socket, used for internal requests (i.e. from providers).
The Coordinator looks for providers of the task in the well defined directory
structure and sends the request to an appropriate provider. If the request has
an http response hook rather than a unix socket, it is tracked and a new request
using the Coordinator's response socket as the hook is sent to the appropriate
task. In this way, responses to external requests first come back to the
Coordinator and then go to the original response hook, rather than Providers
responding directly to the outside world. Similarly, if the response to a
proxied request contains a StreamURL, the Coordinator proxies the stream,
modifying the StreamURL being sent externally appropriately.

### Endpoints

    External Request: http, /
    Internal Request: unix, /[socket_dir]/coordinator/[coordinator name].sock
    Internal Response: unix, /[socket_dir]/response/[coordinator name].sock
    Proxied Stream: http, /stream?addr=[original StreamURL]

### Config

    {
    	"config_file": "/path/to/config/file.json",
    	"socket_dir": "/base/path/for/sockets",
    	"service_name": "NameOfThisCoordinator",
    	"external_port": 8080,
    	"request_timeout": 0,
    	"log_level": "warning"
    }

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

#### func (*Config) ExternalPort

```go
func (c *Config) ExternalPort() int
```
ExternalPort returns the port to listen on for external requests.

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

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type ConfigData

```go
type ConfigData struct {
	SocketDir      string `json:"socket_dir"`
	ServiceName    string `json:"service_name"`
	ExternalPort   uint   `json:"external_port"`
	RequestTimeout uint   `json:"request_timeout"`
	LogLevel       string `json:"log_level"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type Server

```go
type Server struct {
}
```

Server is the coordinator server. It handles accepting internal and external
requests and proxying them to appropriate providers.

#### func  NewServer

```go
func NewServer(config *Config) (*Server, error)
```
NewServer creates and initializes a new instance of Server.

#### func (*Server) Start

```go
func (s *Server) Start() error
```
Start starts the server, running all of the listeners and proxy tracker.

#### func (*Server) Stop

```go
func (s *Server) Stop()
```
Stop stops the server, gracefully stopping all of the listeners and proxy
tracker.

#### func (*Server) StopOnSignal

```go
func (s *Server) StopOnSignal(signals ...os.Signal)
```
StopOnSignal will wait until one of the specified signals is received and then
stop the server. If no signals are specified, it will use a default set.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
