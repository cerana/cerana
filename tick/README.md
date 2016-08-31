# tick

[![tick](https://godoc.org/github.com/cerana/cerana/tick?status.svg)](https://godoc.org/github.com/cerana/cerana/tick)



## Usage

#### func  GetIP

```go
func GetIP(config Configer, tracker *acomm.Tracker) (net.IP, error)
```
GetIP retrieves the ip of the current node, very often needed for ticks.

#### func  RunTick

```go
func RunTick(config Configer, tick ActionFn) (chan struct{}, error)
```
RunTick runs the supplied function on a configured interval. Putting an entry
into the returned chan, as well as sending an os signal, can be used to stop the
tick.

#### type ActionFn

```go
type ActionFn func(Configer, *acomm.Tracker) error
```

ActionFn is a function that can be run on a tick interval.

#### type Config

```go
type Config struct {
}
```

Config is the configuration for a tick.

#### func  NewConfig

```go
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config
```
NewConfig creates a new instance of Config.

#### func (*Config) ClusterDataURL

```go
func (c *Config) ClusterDataURL() *url.URL
```
ClusterDataURL returns the url of the layer 2 coordinator, used for cluster
information.

#### func (*Config) ConfigFile

```go
func (c *Config) ConfigFile() string
```
ConfigFile returns a path to a config file.

#### func (*Config) HTTPResponseURL

```go
func (c *Config) HTTPResponseURL() *url.URL
```
HTTPResponseURL returns the url of the http response listener.

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads the config.

#### func (*Config) LogLevel

```go
func (c *Config) LogLevel() string
```
LogLevel returns the log level.

#### func (*Config) NodeDataURL

```go
func (c *Config) NodeDataURL() *url.URL
```
NodeDataURL returns the url of the layer 1 coordinator, used for node
information.

#### func (*Config) RequestTimeout

```go
func (c *Config) RequestTimeout() time.Duration
```
RequestTimeout returns the default timeout for task requests.

#### func (*Config) SetupLogging

```go
func (c *Config) SetupLogging() error
```
SetupLogging sets up logging with the log level and formatting.

#### func (*Config) TickInterval

```go
func (c *Config) TickInterval() time.Duration
```
TickInterval returns how often the tick function should be executed.

#### func (*Config) TickRetryInterval

```go
func (c *Config) TickRetryInterval() time.Duration
```
TickRetryInterval returns how often the tick function should be executed after
an error.

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate ensures the configuration is valid.

#### func (*Config) ValidateURL

```go
func (c *Config) ValidateURL(name string, required bool) error
```
ValidateURL is used in validation for checking url parameters.

#### type ConfigData

```go
type ConfigData struct {
	NodeDataURL       string `json:"nodeDataURL"`
	ClusterDataURL    string `json:"clusterDataURL"`
	HTTPResponseAddr  string `json:"responseAddr"`
	LogLevel          string `json:"logLevel"`
	RequestTimeout    string `json:"requestTimeout"`
	TickInterval      string `json:"tickInterval"`
	TickRetryInterval string `json:"tickRetryInterval"`
}
```

ConfigData defines the structure of the config data (e.g. in the config file)

#### type Configer

```go
type Configer interface {
	ConfigFile() string
	NodeDataURL() *url.URL
	ClusterDataURL() *url.URL
	HTTPResponseURL() *url.URL
	LogLevel() string
	RequestTimeout() time.Duration
	TickInterval() time.Duration
	TickRetryInterval() time.Duration
}
```

Configer is an interface that provides information the tick needs to run. It
allows more complex configs to be passed through to the tick function.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
