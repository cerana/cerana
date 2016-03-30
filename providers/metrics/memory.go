package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/mem"
)

// MemoryResult is the result for the Memory handler.
type MemoryResult struct {
	Swap    *mem.SwapMemoryStat    `json:"swap"`
	Virtual *mem.VirtualMemoryStat `json:"virtual"`
}

// Memory returns information about the virtual and swap memory.
func (m *Metrics) Memory(req *acomm.Request) (interface{}, *url.URL, error) {
	swap, err := mem.SwapMemory()
	if err != nil {
		return nil, nil, err
	}

	virtual, err := mem.VirtualMemory()
	if err != nil {
		return nil, nil, err
	}

	return &MemoryResult{Swap: swap, Virtual: virtual}, nil, nil
}
