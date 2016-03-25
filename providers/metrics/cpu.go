package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

type CPUResult struct {
	Info  []cpu.CPUInfoStat  `json:"info"`
	Load  *load.LoadAvgStat  `json:"load"`
	Times []cpu.CPUTimesStat `json:"times"`
}

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
