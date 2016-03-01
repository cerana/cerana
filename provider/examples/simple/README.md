# simple

[![simple](https://godoc.org/github.com/mistifyio/provider/examples/simple?status.png)](https://godoc.org/github.com/mistifyio/provider/examples/simple)

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

#### type DelayedRespArgs

```go
type DelayedRespArgs struct {
	Delay time.Duration `json:"delay"`
}
```

DelayedRespArgs are arguments for the DelayedResp handler.

#### type DelayedRespResult

```go
type DelayedRespResult struct {
	Delay       time.Duration `json:"delay"`
	ReceivedAt  time.Time     `json:"received_at"`
	RespondedAt time.Time     `json:"responded_at"`
}
```

DelayedRespResult is the result data for the DelayedResp handler.

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

#### type Simple

```go
type Simple struct {
}
```

Simple is a simple provider implementation.

#### func  NewSimple

```go
func NewSimple(config *provider.Config, tracker *acomm.Tracker) *Simple
```
NewSimple creates a new instance of Simple.

#### func (*Simple) CPUInfo

```go
func (s *Simple) CPUInfo(req *acomm.Request) (interface{}, *url.URL, error)
```
CPUInfo is a task handler to retrieve information about CPUs.

#### func (*Simple) DelayedResp

```go
func (s *Simple) DelayedResp(req *acomm.Request) (interface{}, *url.URL, error)
```
DelayedResp is a task handler that waits a specified time before returning.

#### func (*Simple) DiskInfo

```go
func (s *Simple) DiskInfo(req *acomm.Request) (interface{}, *url.URL, error)
```
DiskInfo is a task handler to retrieve information about disks.

#### func (*Simple) RegisterTasks

```go
func (s *Simple) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Simple's task handlers with the server.

#### func (*Simple) StreamEcho

```go
func (s *Simple) StreamEcho(req *acomm.Request) (interface{}, *url.URL, error)
```
StreamEcho is a task handler to echo input back via streaming data.

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

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
