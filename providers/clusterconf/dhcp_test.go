package clusterconf_test

import (
	"encoding/json"
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

func (s *clusterConf) TestGetDHCP() {
	conf := clusterconf.DHCPConfig{
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
	}

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
		ok   bool
		conf clusterconf.DHCPConfig
	}{
		{desc: "duration is invalid"},
		{desc: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: 25 * time.Hour,
			},
		},
		{desc: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: 30 * time.Minute,
			},
		},
		{desc: "net.IP is required",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
			},
		},
		{desc: "net.IP must be IPv4",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.ParseIP("::1"),
				},
			},
		},
		{desc: "net.IP must not be 0.0.0.0",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.IPv4zero,
				},
			},
		},
		{desc: "net.Mask is required",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Net: net.IPNet{
					IP: net.ParseIP("127.0.0.1"),
				},
			},
		},
		{desc: "gateway is unreachable",
			conf: clusterconf.DHCPConfig{
				Duration: 1 * time.Hour,
				Gateway:  net.IPv4(10, 0, 10, byte(rand.Intn(255))),
				Net: net.IPNet{
					IP:   net.IPv4(10, 100, 10, 0),
					Mask: net.IPMask{255, 255, 255, 0},
				},
			},
		},
		{desc: "", ok: true,
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
		s.Require().NoError(err)

		resp, url, err := s.clusterConf.SetDHCP(req)
		s.Nil(resp)
		s.Nil(url)
		if !t.ok {
			s.EqualError(err, t.desc)
			continue
		}
		if !s.NoError(err) {
			continue
		}

		resp, url, err = s.clusterConf.GetDHCP(nil)
		s.Require().NoError(err)
		s.Nil(url)

		got := resp.(clusterconf.DHCPConfig)
		s.Equal(t.conf, got)
	}
}