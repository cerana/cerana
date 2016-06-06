package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

// CPUResult is the result of the CPU handler.
type CPUResult struct {
	Info  []cpu.InfoStat  `json:"info"`
	Load  load.AvgStat    `json:"load"`
	Times []cpu.TimesStat `json:"times"`
}

// CPU returns information about the CPU hardware, times, and load.
func (m *Metrics) CPU(req *acomm.Request) (interface{}, *url.URL, error) {
	info, err := cpu.Info()
	if err != nil {
		return nil, nil, err
	}

	loadAvg, err := load.Avg()
	if err != nil {
		return nil, nil, err
	}

	times, err := cpu.Times(true)
	if err != nil {
		return nil, nil, err
	}

	return &CPUResult{Info: info, Load: *loadAvg, Times: times}, nil, nil
}
