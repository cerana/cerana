package metrics

import (
	"net/url"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/shirou/gopsutil/host"
)

// Host returns information about the host machine.
func (m *Metrics) Host(req *acomm.Request) (interface{}, *url.URL, error) {
	hostInfo, err := host.Info()
	return hostInfo, nil, errors.Wrap(err)
}
