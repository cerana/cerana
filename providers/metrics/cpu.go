package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

// CPUResult is the result of the CPU handler.
type CPUResult struct {
	Info  []cpu.CPUInfoStat  `json:"info"`
	Load  *load.LoadAvgStat  `json:"load"`
	Times []cpu.CPUTimesStat `json:"times"`
}

// CPU returns information about the CPU hardware, times, and load.
func (m *Metrics) CPU(req *acomm.Request) (interface{}, *url.URL, error) {
	info, err := cpu.CPUInfo()
	if err != nil {
		return nil, nil, err
	}

	loadAvg, err := load.LoadAvg()
	if err != nil {
		return nil, nil, err
	}

	times, err := cpu.CPUTimes(true)
	if err != nil {
		return nil, nil, err
	}

	return &CPUResult{Info: info, Load: loadAvg, Times: times}, nil, nil
}
