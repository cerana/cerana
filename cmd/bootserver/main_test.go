package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"testing"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/krolaw/dhcp4"
	"github.com/stretchr/testify/suite"
)

func uniqueMAC(t *testing.T, macs map[string]struct{}) string {
	for range [5]struct{}{} {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		buf[0] = (buf[0] | 2) & 0xfe
		mac := fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
		if _, ok := macs[mac]; !ok {
			return mac
		}
	}
	t.Fatal(errors.New("failed to create unique mac"))
	panic("")
}

func uniqueIP(t *testing.T, randIP func() net.IP, ips map[string]string) net.IP {
LOOP:
	for range [5]struct{}{} {
		ip := randIP()
		for _, ipS := range ips {
			if ipS == ip.String() {
				continue LOOP
			}
		}
		return ip
	}
	t.Fatal(errors.New("failed to create unique ip"))
	panic("")
}

type DHCPS struct {
	suite.Suite
	dhcp    *dhcp.Mock
	coord   *url.URL
	tracker *acomm.Tracker
}

type iface struct {
	ip   string
	name string
}

func (i iface) Addrs() ([]net.Addr, error) {
	return []net.Addr{i}, nil
}
func (i iface) Network() string {
	return i.name
}
func (i iface) String() string {
	return i.ip
}

func TestDHCP(t *testing.T) {
	suite.Run(t, &DHCPS{})
}

func (s *DHCPS) SetupSuite() {
	coordinator, err := test.NewCoordinator("")
	s.Require().NoError(err)

	sURL := coordinator.NewProviderViper().GetString("coordinator_url")
	s.coord, err = url.Parse(sURL)
	s.Require().NoError(err)

	s.tracker = coordinator.ProviderTracker()

	d := dhcp.NewMock()
	s.dhcp = d

	coordinator.RegisterProvider(d)
	s.Require().NoError(coordinator.Start())
}

func randXID() []byte {
	xid := make([]byte, 4)
	for i := range xid {
		xid[i] = byte(rand.Intn(255))
	}
	return xid
}

func (s *DHCPS) TestDHCP() {
	h := dhcpHandler{
		iface: iface{
			ip:   "10.0.0.1/8",
			name: "eth0",
		},
		tracker:     s.tracker,
		coordinator: s.coord,
	}

	ips := make(map[string]string)
	macs := make(map[string]struct{})

	MAC := uniqueMAC(s.T(), macs)
	macs[MAC] = struct{}{}

	tests := []struct {
		desc    string
		dupe    bool
		ip      net.IP
		mac     string
		nack    bool
		Nil     bool
		op      dhcp4.MessageType
		options dhcp4.Options
		store   bool
	}{
		{desc: "DISCOVER",
			mac: MAC, op: dhcp4.Discover, store: true},
		{desc: "REQUEST: previously offered ip",
			dupe: true, mac: MAC, op: dhcp4.Request},
		{desc: "REQUEST: previously offered ip",
			dupe: true, mac: MAC, op: dhcp4.Request},
		{desc: "DISCOVER: assert uniq ip",
			op: dhcp4.Discover, store: true},
		{desc: "REQUEST: for taken ip",
			dupe: true,
			mac:  uniqueMAC(s.T(), macs),
			nack: true,
			op:   dhcp4.Request,
		},
		{desc: "REQUEST: unallocated ip",
			ip:   uniqueIP(s.T(), s.dhcp.RandIP, ips),
			nack: true,
			op:   dhcp4.Request,
		},
		{desc: "REQUEST: intended for a different server",
			Nil: true,
			op:  dhcp4.Request,
			options: dhcp4.Options{
				dhcp4.OptionServerIdentifier: []byte(uniqueMAC(s.T(), macs)),
			},
		},
		{desc: "REQUEST: IPv6",
			ip:   net.IPv6zero,
			nack: true,
			op:   dhcp4.Request,
		},
		{desc: "REQUEST: zero ip",
			ip:   net.IPv4zero,
			nack: true,
			op:   dhcp4.Request,
		},
		/*
			//TODO check SPEC / compare against dnsmasq/isc-dhcpd
			{desc: "DISCOVER: ip out of range",
				ip:   net.ParseIP("172.16.10.10"),
				nack: true,
				op:   dhcp4.Discover,
			},
			{desc: "REQUEST: different ip",
				ip:  uniqueIP(s.T(), s.dhcp.RandIP, ips),
				mac: MAC, op: dhcp4.Request},
		*/
	}

	for _, t := range tests {
		s.T().Log(t.desc)

		if t.mac == "" {
			t.mac = uniqueMAC(s.T(), macs)
		}
		mac, err := net.ParseMAC(t.mac)
		s.Require().NoError(err)

		_, ok := ips[t.mac]
		var ip net.IP
		switch {
		case t.ip != nil:
			ip = t.ip
		case ok:
			ip = net.ParseIP(ips[t.mac])
		case t.dupe:
			// t.dupe indicates that we want to test an ip collision
			for _, takenIP := range ips {
				ip = net.ParseIP(takenIP)
				break
			}
		}
		req := dhcp4.RequestPacket(t.op, mac, ip, randXID(), true, nil)
		resp := h.ServeDHCP(req, t.op, t.options)
		if t.Nil {
			s.Require().Nil(err)
			continue
		}
		if t.nack {
			s.Require().NotNil(resp)
			mtype := resp.ParseOptions()[dhcp4.OptionDHCPMessageType]
			s.Require().Len(mtype, 1)
			s.Require().Equal(uint8(dhcp4.NAK), mtype[0])
			continue
		}

		ipS, ok := ips[t.mac]
		if t.dupe {
			s.Require().True(ok)
			s.Require().Equal(ipS, resp.YIAddr().String())
		} else {
			s.Require().False(ok)
		}
		if t.store {
			ips[t.mac] = resp.YIAddr().String()
		}
	}

}
