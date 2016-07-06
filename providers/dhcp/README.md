# dhcp

[![dhcp](https://godoc.org/github.com/cerana/cerana/providers/dhcp?status.svg)](https://godoc.org/github.com/cerana/cerana/providers/dhcp)



## Usage

#### type Addresses

```go
type Addresses struct {
	MAC string `json:"mac"`
	IP  string `json:"ip"`
}
```

Addresses specifies the argument to all endpoints

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

#### func (*Config) DNSServers

```go
func (c *Config) DNSServers() []net.IP
```
DNSServers returns the dns server addresses

#### func (*Config) Gateway

```go
func (c *Config) Gateway() net.IP
```
Gateway returns the default gateway address

#### func (*Config) LeaseDuration

```go
func (c *Config) LeaseDuration() time.Duration
```
LeaseDuration returns the client lease duration

#### func (*Config) LoadConfig

```go
func (c *Config) LoadConfig() error
```
LoadConfig loads and validates the KV provider config.

#### func (*Config) Network

```go
func (c *Config) Network() *net.IPNet
```
Network returns the ip range

#### func (*Config) Validate

```go
func (c *Config) Validate() error
```
Validate returns whether the config is valid, containing necessary values.

#### type DHCP

```go
type DHCP struct {
}
```

DHCP is a provider of dhcp functionality.

#### func  New

```go
func New(config *Config, tracker *acomm.Tracker) (*DHCP, error)
```
New creates a new instance of DHCP.

#### func (*DHCP) RegisterTasks

```go
func (d *DHCP) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of DHCP's task handlers with the server.

#### type Lease

```go
type Lease struct {
	Net      net.IPNet     `json:"net"`
	Gateway  net.IP        `json:"gateway"`
	Duration time.Duration `json:"duration"`
	DNS      []net.IP      `json:"dns"`
}
```

Lease specifies the dhcp lease returned from the "dhcp-offer-lease" endpoint.

#### type Mock

```go
type Mock struct {
	sync.Mutex
}
```

Mock is a mock dhcp provider.

#### func  NewMock

```go
func NewMock(config *provider.Config, tracker *acomm.Tracker) (*Mock, error)
```
NewMock creates a new Mock.

#### func (*Mock) Expire

```go
func (m *Mock) Expire(mac string)
```
Expire will remove an entry from memory as if the ephemeral key expired.

#### func (*Mock) RegisterTasks

```go
func (m *Mock) RegisterTasks(server provider.Server)
```
RegisterTasks registers all of Mock's task handlers with the server.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
