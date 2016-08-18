package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
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
		return nil, nil, errors.Wrap(err)
	}

	virtual, err := mem.VirtualMemory()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	return &MemoryResult{Swap: swap, Virtual: virtual}, nil, nil
}
