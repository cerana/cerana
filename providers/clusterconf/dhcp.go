package clusterconf

import (
	"encoding/json"
	"net"
	"net/url"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/errors"
)

const dhcpPrefix string = "dhcp"

// DHCPConfig represents the dhcp settings for the cluster.
type DHCPConfig struct {
	DNS      []string `json:"dns"`
	Duration string   `json:"duration"`
	Gateway  string   `json:"gateway"`
	Net      string   `json:"net"`
}

// Validate validates the DHCPConfig settings.
func (c *DHCPConfig) Validate() error {
	duration, err := time.ParseDuration(c.Duration)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"duration": c.Duration}, "unable to parse duration")
	}
	if duration < 1*time.Hour || duration > 24*time.Hour {
		return errors.Newv("duration is invalid", map[string]interface{}{"duration": duration})
	}

	_, subnet, err := net.ParseCIDR(c.Net)
	if err != nil {
		return errors.Wrapv(err, map[string]interface{}{"net": c.Net}, "unable to parse subnet CIDR")
	}

	if subnet.IP.To4() == nil {
		return errors.Newv("net.ip must be IPv4", map[string]interface{}{"ip": subnet.IP})
	}
	if subnet.IP.Equal(net.IPv4zero) {
		return errors.New("net.ip must not be 0.0.0.0")
	}

	gateway := net.ParseIP(c.Gateway)
	if gateway != nil {
		if !subnet.Contains(gateway) {
			return errors.Newv("gateway is unreachable", map[string]interface{}{"net": subnet, "gateway": gateway})
		}

		for _, dns := range c.DNS {
			ip := net.ParseIP(dns)
			if ip == nil {
				return errors.Newv("failed to parse DNS IP", map[string]interface{}{"ip": dns})
			}
		}
	}
	return nil
}

// GetDHCP retrieves the current cluster DHCP settings.
func (c *ClusterConf) GetDHCP(*acomm.Request) (interface{}, *url.URL, error) {
	value, err := c.kvGet(dhcpPrefix)
	if err != nil {
		return nil, nil, err
	}

	conf := DHCPConfig{}
	if err := json.Unmarshal(value.Data, &conf); err != nil {
		return nil, nil, errors.Wrapv(err, map[string]interface{}{"json": string(value.Data)})
	}

	return conf, nil, nil
}

// SetDHCP updates the cluster DHCP settings.
func (c *ClusterConf) SetDHCP(req *acomm.Request) (interface{}, *url.URL, error) {
	conf := DHCPConfig{}
	if err := req.UnmarshalArgs(&conf); err != nil {
		return nil, nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, nil, err
	}

	_, err := c.kvUpdate(dhcpPrefix, conf, 0)
	if err != nil {
		err = errors.New("dhcp configuration can not be altered")
	}
	return nil, nil, err
}
