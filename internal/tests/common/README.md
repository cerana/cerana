# common

[![common](https://godoc.org/github.com/cerana/cerana/internal/tests/common?status.svg)](https://godoc.org/github.com/cerana/cerana/internal/tests/common)

Package common contains common utilities and suites to be used in other tests

## Usage

#### func  Build

```go
func Build() error
```
Build builds the current go package.

#### func  ConsulMaker

```go
func ConsulMaker(port uint16, dir, prefix string) *exec.Cmd
```
ConsulMaker will create an exec.Cmd to run consul with the given paramaters

#### func  EtcdMaker

```go
func EtcdMaker(port uint16, dir, prefix string) *exec.Cmd
```
EtcdMaker will create an exec.Cmd to run etcd with the given paramaters

#### func  ExitStatus

```go
func ExitStatus(err error) int
```
ExitStatus tries to extract an exit status code from an error.

#### type Cmd

```go
type Cmd struct {
	Cmd *exec.Cmd
	Out *bytes.Buffer
}
```

Cmd wraps an exec.Cmd with monitoring and easy access to output.

#### func  ExecSync

```go
func ExecSync(cmdName string, args ...string) (*Cmd, error)
```
ExecSync runs a command synchronously, waiting for it to complete.

#### func  Start

```go
func Start(cmdName string, args ...string) (*Cmd, error)
```
Start runs a command asynchronously.

#### func (*Cmd) Alive

```go
func (c *Cmd) Alive() bool
```
Alive returns whether the command process is alive or not.

#### func (*Cmd) ExitStatus

```go
func (c *Cmd) ExitStatus() (int, error)
```
ExitStatus returns the exit status code and error for a command. If the command
is still running or in the process of being shut down, the exit code will be 0
and the returned error will be non-nil.

#### func (*Cmd) Stop

```go
func (c *Cmd) Stop() error
```
Stop kills a running command and waits until it exits. The error returned is
from the Kill call, not the error of the exiting command. For the latter, call
c.Err() after c.Stop().

#### func (*Cmd) Wait

```go
func (c *Cmd) Wait() error
```
Wait waits for a command to finish and returns the exit error.

#### type Suite

```go
type Suite struct {
	suite.Suite
	KVDir      string
	KVPrefix   string
	KVPort     uint16
	KVURL      string
	KV         kv.KV
	KVCmd      *exec.Cmd
	KVCmdMaker func(uint16, string, string) *exec.Cmd
	TestPrefix string
	Context    *lochness.Context
}
```

Suite sets up a general test suite with setup/teardown.

#### func (*Suite) DoRequest

```go
func (s *Suite) DoRequest(method, url string, expectedRespCode int, postBodyStruct interface{}, respBody interface{}) *http.Response
```
DoRequest is a convenience method for making an http request and doing basic
handling of the response.

#### func (*Suite) Messager

```go
func (s *Suite) Messager(prefix string) func(...interface{}) string
```
Messager generates a function for creating a string message with a prefix.

#### func (*Suite) NewFWGroup

```go
func (s *Suite) NewFWGroup() *lochness.FWGroup
```
NewFWGroup creates and saves a new FWGroup.

#### func (*Suite) NewFlavor

```go
func (s *Suite) NewFlavor() *lochness.Flavor
```
NewFlavor creates and saves a new Flavor.

#### func (*Suite) NewGuest

```go
func (s *Suite) NewGuest() *lochness.Guest
```
NewGuest creates and saves a new Guest. Creates any necessary resources.

#### func (*Suite) NewHypervisor

```go
func (s *Suite) NewHypervisor() *lochness.Hypervisor
```
NewHypervisor creates and saves a new Hypervisor.

#### func (*Suite) NewHypervisorWithGuest

```go
func (s *Suite) NewHypervisorWithGuest() (*lochness.Hypervisor, *lochness.Guest)
```
NewHypervisorWithGuest creates and saves a new Hypervisor and Guest, with the
Guest added to the Hypervisor.

#### func (*Suite) NewNetwork

```go
func (s *Suite) NewNetwork() *lochness.Network
```
NewNetwork creates and saves a new Netework.

#### func (*Suite) NewSubnet

```go
func (s *Suite) NewSubnet() *lochness.Subnet
```
NewSubnet creates and saves a new Subnet.

#### func (*Suite) NewVLAN

```go
func (s *Suite) NewVLAN() *lochness.VLAN
```
NewVLAN creates and saves a new VLAN.

#### func (*Suite) NewVLANGroup

```go
func (s *Suite) NewVLANGroup() *lochness.VLANGroup
```
NewVLANGroup creates and saves a new VLANGroup.

#### func (*Suite) PrefixKey

```go
func (s *Suite) PrefixKey(key string) string
```
PrefixKey generates an kv key using the set prefix

#### func (*Suite) SetupSuite

```go
func (s *Suite) SetupSuite()
```
SetupSuite runs a new kv instance.

#### func (*Suite) SetupTest

```go
func (s *Suite) SetupTest()
```
SetupTest prepares anything needed per test.

#### func (*Suite) TearDownSuite

```go
func (s *Suite) TearDownSuite()
```
TearDownSuite stops the kv instance and removes all data.

#### func (*Suite) TearDownTest

```go
func (s *Suite) TearDownTest()
```
TearDownTest cleans the kv instance.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
