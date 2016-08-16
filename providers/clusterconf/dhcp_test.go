package clusterconf_test

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/providers/clusterconf"
)

func (s *clusterConf) setupDHCP(config clusterconf.DHCPConfig) {
	buf, err := json.Marshal(config)
	s.Require().NoError(err)
	s.Require().NoError(s.kv.Set("dhcp", string(buf)))
}

type DHCPConfigStrs struct {
	DNS      []string
	Duration string
	Gateway  string
	Net      string
}

func (d DHCPConfigStrs) toConfig() clusterconf.DHCPConfig {
	c := clusterconf.DHCPConfig{
		DNS:     make([]net.IP, len(d.DNS)),
		Gateway: net.ParseIP(d.Gateway),
	}

	for i, dns := range d.DNS {
		c.DNS[i] = net.ParseIP(dns)
	}

	c.Duration, _ = time.ParseDuration(d.Duration)
	_, subnet, _ := net.ParseCIDR(d.Net)
	if subnet != nil {
		c.Net.IP = subnet.IP.To16()
		c.Net.Mask = subnet.Mask
	}

	return c
}

func (s *clusterConf) TestGetDHCP() {
	sconf := DHCPConfigStrs{
		DNS: []string{
			fmt.Sprintf("10.10.1.%d", rand.Intn(255)),
			fmt.Sprintf("10.10.2.%d", rand.Intn(255)),
		},
		Duration: fmt.Sprintf("%dh", 1+rand.Intn(24)),
		Gateway:  fmt.Sprintf("10.10.10.%d", rand.Intn(255)),
		Net:      "10.10.10.0/24",
	}
	conf := sconf.toConfig()

	_, _, err := s.clusterConf.GetDHCP(nil)
	s.Error(err, "expected no dhcp settings")

	s.setupDHCP(conf)

	resp, _, err := s.clusterConf.GetDHCP(nil)
	s.NoError(err, "expected some dhcp settings")

	got := resp.(clusterconf.DHCPConfig)
	s.Equal(conf, got)
}

func (s *clusterConf) TestSetDHCP() {
	s.setupDHCP(clusterconf.DHCPConfig{
		DNS: []net.IP{
			net.IPv4(10, 10, 1, byte(rand.Intn(255))),
			net.IPv4(10, 10, 2, byte(rand.Intn(255))),
		},
		Duration: 5*time.Hour + time.Hour*time.Duration(rand.Intn(11)+1),
		Gateway:  net.IPv4(10, 10, 10, byte(rand.Intn(255))),
		Net: net.IPNet{
			IP:   net.IPv4(10, 10, 10, 0),
			Mask: net.IPMask{255, 255, 255, 0},
		},
	})

	tests := []struct {
		desc string
		err  string
		conf clusterconf.DHCPConfig
	}{
		{desc: "duration too short",
			err: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: 30 * time.Minute,
			},
		},
		{desc: "duration too long",
			err: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: 25 * time.Hour,
			},
		},
		{desc: "missing network",
			err: "net.IP is required",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
			},
		},
		{desc: "IPv6",
			err: "net.IP must be IPv4",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.ParseIP("::1"),
				},
			},
		},
		{desc: "IPv4zero",
			err: "net.IP must not be 0.0.0.0",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.IPv4zero,
				},
			},
		},
		{desc: "missing netmask",
			err: "net.Mask is required",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.ParseIP("127.0.0.1"),
				},
			},
		},
		{desc: "unreachable gateway",
			err: "gateway is unreachable",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Gateway:  net.IPv4(10, 0, 10, byte(rand.Intn(255))),
				Net: net.IPNet{
					IP:   net.IPv4(10, 100, 10, 0),
					Mask: net.IPMask{255, 255, 255, 0},
				},
			},
		},
		{desc: "good",
			conf: clusterconf.DHCPConfig{
				DNS: []net.IP{
					net.IPv4(10, 100, 1, byte(rand.Intn(255))),
					net.IPv4(10, 100, 2, byte(rand.Intn(255))),
				},
				Duration: 1 * time.Hour,
				Gateway:  net.IPv4(10, 100, 10, byte(rand.Intn(255))),
				Net: net.IPNet{
					IP:   net.IPv4(10, 100, 10, 0),
					Mask: net.IPMask{255, 255, 255, 0},
				},
			},
		},
	}

	for _, t := range tests {
		req, err := acomm.NewRequest(acomm.RequestOptions{
			Task: "clusterconf-set-dhcp",
			Args: t.conf,
		})
		s.Require().NoError(err, t.desc)

		resp, url, err := s.clusterConf.SetDHCP(req)
		s.Nil(resp, t.desc)
		s.Nil(url, t.desc)
		if t.err != "" {
			s.Contains(err.Error(), t.err, t.desc)
			continue
		}
		if !s.NoError(err, t.desc) {
			continue
		}

		resp, url, err = s.clusterConf.GetDHCP(nil)
		s.Require().NoError(err, t.desc)
		s.Nil(url)

		got := resp.(clusterconf.DHCPConfig)
		s.Equal(t.conf, got, t.desc)
	}
}
