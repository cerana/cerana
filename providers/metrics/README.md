# metrics

[![metrics](https://godoc.org/github.com/cerana/cerana/providers/metrics?status.png)](https://godoc.org/github.com/cerana/cerana/providers/metrics)



## Usage

#### type CPUResult

```go
type CPUResult struct {
	Info  []cpu.InfoStat  `json:"info"`
	Load  *load.AvgStat   `json:"load"`
	Times []cpu.TimesStat `json:"times"`
}
```

CPUResult is the result of the CPU handler.

#### type DiskResult

```go
type DiskResult struct {
	IO         map[string]disk.IOCountersStat `json:"io"`
	Partitions []disk.PartitionStat           `json:"partitions"`
	Usage      []*disk.UsageStat              `json:"usage"`
}
```

DiskResult is the result for the Disk handler.

#### type MemoryResult

```go
type MemoryResult struct {
	Swap    *mem.SwapMemoryStat    `json:"swap"`
	Virtual *mem.VirtualMemoryStat `json:"virtual"`
}
```

MemoryResult is the result for the Memory handler.

#### type Metrics

```go
type Metrics struct{}
```

Metrics is a provider of system info and metrics functionality.

#### func (*Metrics) CPU

```go
func (m *Metrics) CPU(req *acomm.Request) (interface{}, *url.URL, error)
```
CPU returns information about the CPU hardware, times, and load.

#### func (*Metrics) Disk

```go
func (m *Metrics) Disk(req *acomm.Request) (interface{}, *url.URL, error)
```
Disk returns information about the disk partitions, io, and usage.

#### func (*Metrics) Hardware

```go
func (m *Metrics) Hardware(req *acomm.Request) (interface{}, *url.URL, error)
```
Hardware returns information about the hardware.

#### func (*Metrics) Host

```go
func (m *Metrics) Host(req *acomm.Request) (interface{}, *url.URL, error)
```
Host returns information about the host machine.

#### func (*Metrics) Memory

```go
func (m *Metrics) Memory(req *acomm.Request) (interface{}, *url.URL, error)
```
Memory returns information about the virtual and swap memory.

#### func (*Metrics) Network

```go
func (m *Metrics) Network(req *acomm.Request) (interface{}, *url.URL, error)
```
Network returns information about the net interfaces and io.

#### func (*Metrics) RegisterTasks

```go
func (m *Metrics) RegisterTasks(server *provider.Server)
```
RegisterTasks registers all of Metric's task handlers with the server.

#### type NetworkResult

```go
type NetworkResult struct {
	IO         []net.IOCountersStat `json:"io"`
	Interfaces []net.InterfaceStat  `json:"interfaces"`
}
```

NetworkResult is the result for the Network handler.

--
*Generated with [godocdown](https://github.com/robertkrimen/godocdown)*
