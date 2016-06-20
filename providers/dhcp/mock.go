package dhcp

import (
	"errors"
	"math/rand"
	"net"
	"net/url"
	"sync"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	"github.com/krolaw/dhcp4"
)

// Mock is a mock dhcp provider.
type Mock struct {
	randIP func() net.IP
	sync.Mutex
	macs map[string]string
	ips  map[string]string
}

// NewMock creates a new Mock.
func NewMock(config *provider.Config, tracker *acomm.Tracker) (*Mock, error) {
	_, network, _ := net.ParseCIDR("10.0.0.0/8")
	_, zeros := network.Mask.Size()
	size := (1 << uint(zeros)) - 2
	return &Mock{
		randIP: func() net.IP {
			num := rand.Intn(size)
			return dhcp4.IPAdd(network.IP, num)
		},
		macs: map[string]string{},
		ips:  map[string]string{},
	}, nil
}

// RegisterTasks registers all of Mock's task handlers with the server.
func (m *Mock) RegisterTasks(server provider.Server) {
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
	lease := Lease{}

	m.Lock()
	defer m.Unlock()

	ip := m.ips[addrs.MAC]

	if addrs.IP != "" {
		if addrs.IP == ip {
			lease.Net.IP = net.ParseIP(ip)
			return lease, nil, nil
		}
		// otherwise that ip is alloc'ed for a different client
	}

	// client asked for a different ip than previous alloc'ed
	if ip != "" {
		lease.Net.IP = net.ParseIP(ip)
		return lease, nil, nil
	}

	for range [5]struct{}{} {
		netIP := m.randIP()
		if m.macs[netIP.String()] == "" {
			m.macs[netIP.String()] = addrs.MAC
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
	if ip == "" || ip != mac {
		return nil, nil, errors.New("unknown mac address")
	}

	return nil, nil, nil
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
