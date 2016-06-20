package dhcp

import (
	"errors"
	"net"
	"time"

	"github.com/cerana/cerana/provider"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the provider.
type Config struct {
	*provider.Config
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{Config: provider.NewConfig(flagSet, v)}
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	dur, err := c.LeaseDuration()
	if err != nil {
		return err
	}
	if dur < 1*time.Hour || dur > 24*time.Hour {
		return errors.New("lease must be 1-24 hours")
	}

	g, err := c.Gateway()
	if err != nil {
		return err
	}

	net, err := c.Network()
	if err != nil {
		return err
	}

	if !net.Contains(g) {
		return errors.New("gateway is not reachable from subnet")
	}

	return c.Config.Validate()
}

// LoadConfig loads and validates the KV provider config.
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}

// LeaseDuration returns the client lease duration
func (c *Config) LeaseDuration() (time.Duration, error) {
	var lease time.Duration
	if err := c.UnmarshalKey("leaseDuration", &lease); err != nil {
		return 0, err
	}

	return lease, nil
}

// Gateway returns the default gateway address
func (c *Config) Gateway() (net.IP, error) {
	var address string
	if err := c.UnmarshalKey("gateway", &address); err != nil {
		return nil, err
	}

	ip := net.ParseIP(address)
	if ip == nil {
		return nil, errors.New("gateway is not a valid ip address")
	}

	return ip, nil
}

// DNSServers returns the dns server addresses
func (c *Config) DNSServers() ([]net.IP, error) {
	var addresses []string
	if err := c.UnmarshalKey("dnsServers", &addresses); err != nil {
		return nil, err
	}

	ips := make([]net.IP, len(addresses))
	for i := range addresses {
		ip := net.ParseIP(addresses[i])
		if ip == nil {
			return nil, errors.New("invalid dns server ip address")
		}
		ips[i] = ip
	}

	return ips, nil
}

// Network returns the ip range
func (c *Config) Network() (*net.IPNet, error) {
	var address string
	if err := c.UnmarshalKey("network", &address); err != nil {
		return nil, err
	}

	_, net, err := net.ParseCIDR(address)
	if err != nil {
		return nil, err
	}

	return net, nil
}
