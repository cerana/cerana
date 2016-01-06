# simple

[![simple](https://godoc.org/github.com/mistifyio/provider-simple?status.png)](https://godoc.org/github.com/mistifyio/provider-simple)

Package simple is a simple task provider proof of concept.

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

#### func (*Config) LoadConfigFile

```go
func (c *Config) LoadConfigFile() error
```
LoadConfigFile attempts to load a config file.

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

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
