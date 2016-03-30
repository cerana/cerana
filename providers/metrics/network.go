package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/net"
)

// NetworkResult is the result for the Network handler.
type NetworkResult struct {
	IO         []net.NetIOCountersStat `json:"io"`
	Interfaces []net.NetInterfaceStat  `json:"interfaces"`
}

// Network returns information about the net interfaces and io.
func (m *Metrics) Network(req *acomm.Request) (interface{}, *url.URL, error) {
	interfaces, err := net.NetInterfaces()
	if err != nil {
		return nil, nil, err
	}

	io, err := net.NetIOCounters(true)
	if err != nil {
		return nil, nil, err
	}

	return &NetworkResult{IO: io, Interfaces: interfaces}, nil, nil
}
