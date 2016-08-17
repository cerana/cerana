package clusterconf_test

import (
	"encoding/json"
	"fmt"
	"math/rand"

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
		DNS: []string{
			fmt.Sprintf("10.10.1.%d", rand.Intn(255)),
			fmt.Sprintf("10.10.2.%d", rand.Intn(255)),
		},
		Duration: fmt.Sprintf("%dh", 1+rand.Intn(24)),
		Gateway:  fmt.Sprintf("10.10.10.%d", rand.Intn(255)),
		Net:      "10.10.10.0/24",
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
	tests := []struct {
		desc string
		err  string
		conf clusterconf.DHCPConfig
	}{
		{desc: "duration too short",
			err: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: "30m",
			},
		},
		{desc: "duration too long",
			err: "duration is invalid",
			conf: clusterconf.DHCPConfig{
				Duration: "25h",
			},
		},
		{desc: "missing network",
			err: "invalid CIDR address: ",
			conf: clusterconf.DHCPConfig{
				Duration: "1h",
			},
		},
		{desc: "IPv6",
			err: "net.ip must be IPv4",
			conf: clusterconf.DHCPConfig{
				Duration: "1h",
				Net:      "::1/128",
			},
		},
		{desc: "IPv4zero",
			err: "net.ip must not be 0.0.0.0",
			conf: clusterconf.DHCPConfig{
				Duration: "1h",
				Net:      "0.0.0.0/0",
			},
		},
		{desc: "missing netmask",
			err: "invalid CIDR address: 10.100.10.0",
			conf: clusterconf.DHCPConfig{
				Duration: "1h",
				Net:      "10.100.10.0",
			},
		},
		{desc: "unreachable gateway",
			err: "gateway is unreachable",
			conf: clusterconf.DHCPConfig{
				Duration: "1h",
				Gateway:  fmt.Sprintf("10.0.10.%d", rand.Intn(255)),
				Net:      "10.100.10.0/24",
			},
		},
		{desc: "good",
			conf: clusterconf.DHCPConfig{
				DNS: []string{
					fmt.Sprintf("10.100.1.%d", rand.Intn(255)),
					fmt.Sprintf("10.100.2.%d", rand.Intn(255)),
				},
				Duration: "1h",
				Gateway:  fmt.Sprintf("10.100.10.%d", rand.Intn(255)),
				Net:      "10.100.10.0/24",
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
