# test

[![test](https://godoc.org/github.com/cerana/cerana/pkg/test?status.svg)](https://godoc.org/github.com/cerana/cerana/pkg/test)



## Usage

#### type Coordinator

```go
type Coordinator struct {
	SocketDir string
}
```

Coordinator holds a coordinator server and a provider server with one or more
registered mock Providers to be used for testing.

#### func  NewCoordinator

```go
func NewCoordinator(baseDir string) (*Coordinator, error)
```
NewCoordinator creates a new Coordinator. The coordinator sever will be given a
temporary socket directory and external port.

#### func (*Coordinator) Cleanup

```go
func (c *Coordinator) Cleanup() error
```
Cleanup removes the temporary socket directory.

#### func (*Coordinator) NewProviderViper

```go
func (c *Coordinator) NewProviderViper() *viper.Viper
```
NewProviderViper prepares a basic viper instance for a Provider, setting
appropriate values corresponding to the coordinator and provider server.

#### func (*Coordinator) ProviderTracker

```go
func (c *Coordinator) ProviderTracker() *acomm.Tracker
```
ProviderTracker returns the tracker of the provider server.

#### func (*Coordinator) RegisterProvider

```go
func (c *Coordinator) RegisterProvider(p provider.Provider)
```
RegisterProvider registers a Provider's tasks with the internal Provider server.

#### func (*Coordinator) Start

```go
func (c *Coordinator) Start() error
```
Start starts the Coordinator and Provider servers.

#### func (*Coordinator) Stop

```go
func (c *Coordinator) Stop()
```
Stop stops the Coordinator and Provider servers.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
