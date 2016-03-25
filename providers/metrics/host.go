package metrics

import (
	"net/url"

	"github.com/mistifyio/mistify/acomm"
	"github.com/shirou/gopsutil/host"
)

func (m *Metrics) Host(req *acomm.Request) (interface{}, *url.URL, error) {
	hostInfo, err := host.HostInfo()
	return hostInfo, nil, err
}
