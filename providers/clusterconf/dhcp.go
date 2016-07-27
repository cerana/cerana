package clusterconf

import (
	"encoding/json"
	"errors"
	"net"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
)

const dhcpPrefix string = "dhcp"

// DHCPConfig represents the dhcp settings for the cluster.
type DHCPConfig struct {
	DNS      []net.IP      `json:"dns"`
	Duration time.Duration `json:"duration"`
	Gateway  net.IP        `json:"gateway"`
	Net      net.IPNet     `json:"net"`
}

// Validate validates the DHCPConfig settings.
func (c *DHCPConfig) Validate() error {
	if c.Duration < 1*time.Hour || c.Duration > 24*time.Hour {
		return errors.New("duration is invalid")
	}
	if c.Net.IP == nil {
		return errors.New("net.IP is required")
	}
	c.Net.IP = c.Net.IP.To4()
	if c.Net.IP == nil {
		return errors.New("net.IP must be IPv4")
	}
	if c.Net.IP.Equal(net.IPv4zero) {
		return errors.New("net.IP must not be 0.0.0.0")
	}
	if c.Net.Mask == nil {
		return errors.New("net.Mask is required")
	}
	if c.Gateway != nil {
		if !c.Net.Contains(c.Gateway) {
			return errors.New("gateway is unreachable")
		}
	}
	return nil
}

// GetDHCP retrieves the current cluster DHCP settings.
func (c *ClusterConf) GetDHCP(*acomm.Request) (interface{}, *url.URL, error) {
	value, err := c.kvGet(dhcpPrefix)
	if err != nil {
	}

	config := DHCPConfig{}
	if err := json.Unmarshal(value.Data, &config); err != nil {
		return nil, nil, err
	}
	return config, nil, nil
}

// SetDHCP updates the cluster DHCP settings.
func (c *ClusterConf) SetDHCP(req *acomm.Request) (interface{}, *url.URL, error) {
	conf := &DHCPConfig{}
	if err := req.UnmarshalArgs(conf); err != nil {
		return nil, nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, nil, err
	}

	index := uint64(0)
	value, err := c.kvGet(dhcpPrefix)
	if err == nil {
		index = value.Index
	}

	_, err = c.kvUpdate(dhcpPrefix, conf, index)
	return nil, nil, err
}
