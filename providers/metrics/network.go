package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/shirou/gopsutil/net"
)

// NetworkResult is the result for the Network handler.
type NetworkResult struct {
	IO         []net.IOCountersStat `json:"io"`
	Interfaces []net.InterfaceStat  `json:"interfaces"`
}

// Network returns information about the net interfaces and io.
func (m *Metrics) Network(req *acomm.Request) (interface{}, *url.URL, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	io, err := net.IOCounters(true)
	if err != nil {
		return nil, nil, errors.Wrap(err)
	}

	return &NetworkResult{IO: io, Interfaces: interfaces}, nil, nil
}
