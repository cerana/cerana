package dhcp

import (
	"fmt"
	"math/rand"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/kv"
	"github.com/krolaw/dhcp4"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type DHCPS struct {
	suite.Suite
	coord  *test.Coordinator
	dhcp   *DHCP
	config *Config
	kv     *kv.Mock
}

func randMAC(t *testing.T) string {
	buf := make([]byte, 6)
	_, err := rand.Read(buf)
	if err != nil {
		t.Error(err)
	}
	buf[0] = (buf[0] | 2) & 0xfe
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
}

func TestDHCP(t *testing.T) {
	suite.Run(t, &DHCPS{})
}

func (s *DHCPS) SetupSuite() {
	coordinator, err := test.NewCoordinator("")
	s.Require().NoError(err)
	s.coord = coordinator

	v := coordinator.NewProviderViper()
	v.Set("lease-duration", 24*time.Hour)
	v.Set("dns-servers", []string{"10.0.0.1", "10.0.0.2"})
	v.Set("gateway", "10.0.0.1")
	v.Set("network", "10.0.0.1/24")

	flagset := pflag.NewFlagSet("dhcp", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.config = &Config{Config: config, viper: v}
	s.Require().NoError(s.config.LoadConfig())
	s.Require().NoError(s.config.SetupLogging())
	s.Require().NoError(flagset.Parse([]string{}))

	tracker, err := acomm.NewTracker(filepath.Join(coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(tracker.Start())

	mock, err := kv.NewMock(config, tracker)
	s.Require().NoError(err)
	s.kv = mock

	coordinator.RegisterProvider(mock)
	s.Require().NoError(coordinator.Start())

	dhcp, err := New(s.config, tracker)
	s.Require().NoError(err)
	s.dhcp = dhcp
}

func (s *DHCPS) TearDownTest() {
	_ = s.kv.Clean(prefix)
}

func (s *DHCPS) TearDownSuite() {
	s.kv.Stop()
	s.coord.Stop()
}

func (s *DHCPS) TestConfig() {
	t := []struct {
		name     string
		getter   func(*Config) interface{}
		expected interface{}
	}{
		{"lease duration",
			func(c *Config) interface{} { return c.LeaseDuration() },
			24 * time.Hour,
		},
		{"gateway",
			func(c *Config) interface{} { return c.Gateway() },
			net.ParseIP("10.0.0.1"),
		},
		{"network",
			func(c *Config) interface{} { return c.Network() },
			func() *net.IPNet { _, network, _ := net.ParseCIDR("10.0.0.1/24"); return network }(),
		},
		{"dns servers",
			func(c *Config) interface{} { return c.DNSServers() },
			[]net.IP{net.ParseIP("10.0.0.1"), net.ParseIP("10.0.0.2")},
		},
	}

	for _, t := range t {
		got := t.getter(s.config)
		s.NotNil(got)
		s.Equal(t.expected, got, t.name)
	}
}

func (s *DHCPS) TestGetAddressBasic() {
	mac := randMAC(s.T())

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "dhcp-offer-lease",
		Args: Addresses{
			MAC: mac,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(req)

	resp, url, err := s.dhcp.get(req)
	s.Require().Nil(err)
	s.Require().Nil(url)
	s.Require().NotNil(resp)

	lease := resp.(Lease)
	s.Require().Equal(s.dhcp.config.LeaseDuration(), lease.Duration)
	s.Require().Equal(s.dhcp.config.Gateway(), lease.Gateway)
	s.Require().Equal(s.dhcp.config.Network().Mask, lease.Net.Mask)
	s.Require().Equal(s.dhcp.config.Network().IP, lease.Net.IP.Mask(lease.Net.Mask))

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "dhcp-ack-lease",
		Args: Addresses{
			MAC: mac,
			IP:  lease.Net.IP.String(),
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(req)

	resp, url, err = s.dhcp.ack(req)
	s.Require().Nil(err)
	s.Require().Nil(url)
	s.Require().Nil(resp)
}

func (s *DHCPS) TestGetAlmostFull() {
	size, bits := s.dhcp.config.Network().Mask.Size()
	numIPs := (1 << uint(bits-size)) - 2

	const macFormat = "00:ba:dd:be:ef:%02x"
	n := rand.Intn(numIPs) + 1

	wantIP := dhcp4.IPAdd(s.dhcp.config.Network().IP, n)
	wantMAC := fmt.Sprintf(macFormat, n)
	for i := 1; i < numIPs+1; i++ {
		ip := dhcp4.IPAdd(s.dhcp.config.Network().IP, i)
		if ip.String() == wantIP.String() {
			continue
		}
		s.Require().NoError(s.kv.Set(prefix+ip.String(), fmt.Sprintf(macFormat, i)))
	}
	time.Sleep(1 * time.Second)

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "dhcp-offer-lease",
		Args: Addresses{
			MAC: wantMAC,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(req)

	resp, url, err := s.dhcp.get(req)
	s.Require().Nil(err)
	s.Require().Nil(url)
	s.Require().NotNil(resp)

	lease := resp.(Lease)
	s.Require().Equal(wantIP.String(), lease.Net.IP.String())
}

func (s *DHCPS) TestGetTaken() {
	mac := randMAC(s.T())

	req, err := acomm.NewRequest(acomm.RequestOptions{
		Task: "dhcp-offer-lease",
		Args: Addresses{
			MAC: mac,
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(req)

	resp, url, err := s.dhcp.get(req)
	s.Require().Nil(err)
	s.Require().Nil(url)
	s.Require().NotNil(resp)

	lease := resp.(Lease)
	s.Require().Equal(s.dhcp.config.LeaseDuration(), lease.Duration)
	s.Require().Equal(s.dhcp.config.Gateway(), lease.Gateway)
	s.Require().Equal(s.dhcp.config.Network().Mask, lease.Net.Mask)
	s.Require().Equal(s.dhcp.config.Network().IP, lease.Net.IP.Mask(lease.Net.Mask))

	req, err = acomm.NewRequest(acomm.RequestOptions{
		Task: "dhcp-offer-lease",
		Args: Addresses{
			MAC: randMAC(s.T()),
			IP:  lease.Net.IP.String(),
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(req)

	resp, url, err = s.dhcp.get(req)
	s.Require().Nil(err)
	s.Require().Nil(url)
	s.Require().NotNil(resp)

	lease2 := resp.(Lease)
	s.Require().Equal(s.dhcp.config.LeaseDuration(), lease2.Duration)
	s.Require().Equal(s.dhcp.config.Gateway(), lease2.Gateway)
	s.Require().Equal(s.dhcp.config.Network().Mask, lease2.Net.Mask)
	s.Require().Equal(s.dhcp.config.Network().IP, lease2.Net.IP.Mask(lease2.Net.Mask))
	s.Require().NotEqual(lease.Net.IP, lease2.Net.IP)
}

func (s *DHCPS) TestNextGetter() {
	min := uint32(0)
	max := uint32(7)
	tests := []struct {
		name string
		ips  []uint32
		want []uint32
	}{
		{"empty first ok",
			nil, []uint32{1}},
		{"empty first taken",
			nil, []uint32{1, 2}},
		{"first is skipped",
			[]uint32{2, 3}, []uint32{1}},
		{"first is skipped but already taken",
			[]uint32{2, 3}, []uint32{1, 4, 5}},
		{"alternating missed",
			[]uint32{2, 4, 6}, []uint32{1, 3, 5}},
		{"last one",
			[]uint32{1, 2, 3, 4, 5}, []uint32{6}},
		{"non available",
			[]uint32{1, 2, 3, 4, 5, 6}, []uint32{}},
		{"stop correctly",
			[]uint32{1, 2, 3, 4, 5, 6, 7}, []uint32{}},
	}

	for _, t := range tests {
		closer := make(chan struct{})
		got := []uint32{}
		i := 0
		for ip := range nextGetter(closer, t.ips, min, max) {
			got = append(got, ip)

			i++
			if i >= len(t.want) {
				break
			}
		}
		close(closer)
		s.Equal(t.want, got, t.name)
	}
}
