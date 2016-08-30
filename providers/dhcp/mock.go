package dhcp

import (
	"math/rand"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/provider"
	"github.com/krolaw/dhcp4"
	"github.com/spf13/viper"
)

// Mock is a mock dhcp provider.
type Mock struct {
	RandIP func() net.IP
	Config *Config
	sync.Mutex
	macs map[string]string
	ips  map[string]string
}

// NewMock creates a new Mock.
func NewMock() *Mock {
	v := viper.New()
	v.Set("dns-servers", []string{"8.8.8.8", "8.8.4.4"})
	v.Set("gateway", "10.0.0.1")
	v.Set("lease-duration", 24*time.Hour)
	v.Set("network", "10.0.0.1/8")
	conf := NewConfig(nil, v)

	_, network, _ := net.ParseCIDR("10.0.0.0/8")
	ones, bits := network.Mask.Size()
	size := 1<<uint(bits-ones) - 2
	return &Mock{
		Config: conf,
		RandIP: func() net.IP {
			num := rand.Intn(size) + 1
			return dhcp4.IPAdd(network.IP, num)
		},
		macs: map[string]string{},
		ips:  map[string]string{},
	}
}

// RegisterTasks registers all of Mock's task handlers with the server.
func (m *Mock) RegisterTasks(server *provider.Server) {
	server.RegisterTask("dhcp-offer-lease", m.get)
	server.RegisterTask("dhcp-ack-lease", m.ack)
	server.RegisterTask("dhcp-remove-lease", m.remove)
}

func (m *Mock) get(req *acomm.Request) (interface{}, *url.URL, error) {
	addrs := Addresses{}
	if err := req.UnmarshalArgs(&addrs); err != nil {
		return nil, nil, err
	}
	if addrs.MAC == "" {
		return nil, nil, errors.New("missing arg: mac")
	}
	lease := Lease{
		DNS:      m.Config.DNSServers(),
		Duration: m.Config.LeaseDuration(),
		Gateway:  m.Config.Gateway(),
		Net: net.IPNet{
			Mask: m.Config.Network().Mask,
		},
	}

	m.Lock()
	defer m.Unlock()

	ip := m.ips[addrs.MAC]

	if addrs.IP != "" {
		if addrs.IP == ip {
			lease.Net.IP = net.ParseIP(ip)
			return lease, nil, nil
		}

		// client asked for a different ip than previous alloc'ed
		if ip != "" && m.macs[ip] == "" && m.Config.ipInRange(net.ParseIP(ip)) {
			m.macs[ip] = addrs.MAC
			m.ips[addrs.MAC] = ip

			lease.Net.IP = net.ParseIP(ip)
			return lease, nil, nil
		}
	} else if ip != "" {
		lease.Net.IP = net.ParseIP(ip)
		return lease, nil, nil
	}

	for range [5]struct{}{} {
		netIP := m.RandIP()
		if m.macs[netIP.String()] == "" {
			m.macs[netIP.String()] = addrs.MAC
			m.ips[addrs.MAC] = netIP.String()

			lease.Net.IP = netIP
			return lease, nil, nil
		}
	}

	return nil, nil, errors.New("no ip available")
}

func (m *Mock) ack(req *acomm.Request) (interface{}, *url.URL, error) {
	addrs := Addresses{}
	if err := req.UnmarshalArgs(&addrs); err != nil {
		return nil, nil, err
	}
	if addrs.MAC == "" {
		return nil, nil, errors.New("missing arg: mac")
	}
	if addrs.IP == "" {
		return nil, nil, errors.New("missing arg: ip")
	}

	m.Lock()
	defer m.Unlock()

	mac := m.macs[addrs.IP]
	ip := m.ips[addrs.MAC]

	if mac == "" {
		return nil, nil, errors.New("unallocated ip address")
	}
	if ip == "" || m.macs[ip] != mac {
		return nil, nil, errors.New("unknown mac address")
	}

	lease := Lease{
		DNS:      m.Config.DNSServers(),
		Duration: m.Config.LeaseDuration(),
		Gateway:  m.Config.Gateway(),
		Net: net.IPNet{
			IP:   net.ParseIP(ip),
			Mask: m.Config.Network().Mask,
		},
	}
	return lease, nil, nil
}

// Expire will remove an entry from memory as if the ephemeral key expired.
func (m *Mock) Expire(mac string) {

	m.Lock()
	defer m.Unlock()

	delete(m.macs, m.ips[mac])
	delete(m.ips, mac)
}

func (m *Mock) remove(req *acomm.Request) (interface{}, *url.URL, error) {
	addrs := Addresses{}
	if err := req.UnmarshalArgs(&addrs); err != nil {
		return nil, nil, err
	}
	if addrs.MAC == "" {
		return nil, nil, errors.New("missing arg: mac")
	}

	m.Expire(addrs.MAC)

	return nil, nil, nil
}
