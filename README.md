# coordinator

[![coordinator](https://godoc.org/github.com/mistifyio/coordinator?status.png)](https://godoc.org/github.com/mistifyio/coordinator)



## Usage

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

#### func (*Config) ExternalPort

```go
func (c *Config) ExternalPort() int
```
ExternalPort returns the port to listen on for external requests.

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

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

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
