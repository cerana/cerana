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

func validateDHCPConf(conf *DHCPConfig) error {
	if conf.Duration == 0 {
		return errors.New("missing arg: duration")
	}
	if conf.Net.IP == nil {
		return errors.New("missing arg: net.IP")
	}
	conf.Net.IP = conf.Net.IP.To4()
	if conf.Net.IP == nil {
		return errors.New("invalid arg: net.IP must be IPv4")
	}
	if conf.Net.IP.Equal(net.IPv4zero) {
		return errors.New("invalid arg: net.IP must not be 0.0.0.0")
	}
	if conf.Net.Mask == nil {
		return errors.New("missing arg: net.Mask")
	}
	return nil
}

// SetDHCP updates the cluster DHCP settings.
func (c *ClusterConf) SetDHCP(req *acomm.Request) (interface{}, *url.URL, error) {
	conf := DHCPConfig{}
	if err := req.UnmarshalArgs(&conf); err != nil {
		return nil, nil, err
	}

	if err := validateDHCPConf(&conf); err != nil {
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
