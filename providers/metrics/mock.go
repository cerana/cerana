package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
)

// MockMetrics is a mock Metrics provider.
type MockMetrics struct {
	Data *MockMetricsData
}

// MockMetricsData is the in-memory data structure for the MockMetrics.
type MockMetricsData struct {
	CPU      *CPUResult
	Disk     *DiskResult
	Hardware interface{}
	Host     *host.InfoStat
	Memory   *MemoryResult
	Network  *NetworkResult
}

// NewMockMetrics creates a new MockMetrics and populates default data.
func NewMockMetrics() *MockMetrics {
	return &MockMetrics{
		Data: &MockMetricsData{
			CPU: &CPUResult{
				Info: []cpu.InfoStat{
					{
						CPU:        0,
						VendorID:   "",
						Family:     "6",
						Model:      "69",
						Stepping:   1,
						PhysicalID: "0",
						CoreID:     "0",
						Cores:      1,
						ModelName:  "Intel(R) Core(TM) i5-4278U CPU @ 2.60GHz",
						Mhz:        2599.79,
						CacheSize:  3072,
						Flags:      []string{},
					}, {
						CPU:        1,
						VendorID:   "",
						Family:     "6",
						Model:      "69",
						Stepping:   1,
						PhysicalID: "0",
						CoreID:     "1",
						Cores:      1,
						ModelName:  "Intel(R) Core(TM) i5-4278U CPU @ 2.60GHz",
						Mhz:        2599.79,
						CacheSize:  3072,
						Flags:      []string{},
					},
				},
				Load: load.AvgStat{
					Load1:  0.72,
					Load5:  0.22,
					Load15: 0.11,
				},
				Times: []cpu.TimesStat{
					{
						CPU:       "cpu0",
						User:      4973.34,
						System:    3579.08,
						Idle:      824805.64,
						Nice:      0.93,
						Iowait:    921.25,
						Irq:       0,
						Softirq:   674.86,
						Steal:     0,
						Guest:     0,
						GuestNice: 0,
						Stolen:    0,
					}, {
						CPU:       "cpu1",
						User:      5104.66,
						System:    3458.4,
						Idle:      837419.2,
						Nice:      1.29,
						Iowait:    869.55,
						Irq:       0,
						Softirq:   161.14,
						Steal:     0,
						Guest:     0,
						GuestNice: 0,
						Stolen:    0,
					},
				},
			},
			Disk: &DiskResult{
				IO:         make(map[string]disk.IOCountersStat),
				Partitions: []disk.PartitionStat{},
				Usage: []*disk.UsageStat{
					{
						Path:        "/",
						Fstype:      "zfs",
						Total:       100000,
						Free:        50000,
						Used:        50000,
						UsedPercent: 0.5,
					},
				},
			},
			Hardware: struct{}{},
			Host: &host.InfoStat{
				Hostname:       "mockMetrics",
				Uptime:         uint64(123456),
				BootTime:       uint64(123456),
				Procs:          5,
				OS:             "linux",
				Platform:       "cerana",
				PlatformFamily: "nix",
			},
			Memory: &MemoryResult{
				Swap: &mem.SwapMemoryStat{
					Total:       100000,
					Used:        50000,
					Free:        50000,
					UsedPercent: 0.5,
				},
				Virtual: &mem.VirtualMemoryStat{
					Total:       100000,
					Available:   100000,
					Used:        50000,
					Free:        50000,
					UsedPercent: 0.5,
				},
			},
			Network: &NetworkResult{
				IO: []net.IOCountersStat{
					net.IOCountersStat{
						Name:        "fooIface",
						BytesSent:   12345,
						BytesRecv:   54321,
						PacketsSent: 12345,
						PacketsRecv: 54321,
					},
				},
				Interfaces: []net.InterfaceStat{
					net.InterfaceStat{
						Name:         "fooIface",
						MTU:          1500,
						HardwareAddr: "en0",
						Addrs: []net.InterfaceAddr{
							net.InterfaceAddr{"123.123.123.123"},
						},
					},
				},
			},
		},
	}
}

// RegisterTasks registes all MockMetric task handlers.
func (m *MockMetrics) RegisterTasks(server *provider.Server) {
	server.RegisterTask("metrics-cpu", m.CPU)
	server.RegisterTask("metrics-disk", m.Disk)
	server.RegisterTask("metrics-host", m.Host)
	server.RegisterTask("metrics-memory", m.Memory)
	server.RegisterTask("metrics-network", m.Network)
}

func (m *MockMetrics) CPU(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.CPU, nil, nil
}

func (m *MockMetrics) Disk(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.Disk, nil, nil
}

func (m *MockMetrics) Hardware(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.Hardware, nil, nil
}

func (m *MockMetrics) Host(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.Host, nil, nil
}

func (m *MockMetrics) Memory(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.Memory, nil, nil
}

func (m *MockMetrics) Network(req *acomm.Request) (interface{}, *url.URL, error) {
	return m.Data.Network, nil, nil
}
