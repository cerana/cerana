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
	viper *viper.Viper
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{
		Config: provider.NewConfig(flagSet, v),
		viper:  v,
	}
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	dur := c.LeaseDuration()
	if dur < 1*time.Hour || dur > 24*time.Hour {
		return errors.New("lease must be 1-24 hours")
	}

	net, err := c.network()
	if err != nil {
		return err
	}

	g := c.Gateway()
	if g == nil {
		return errors.New("gateway is invalid")
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
func (c *Config) LeaseDuration() time.Duration {
	return c.viper.GetDuration("lease-duration")
}

// Gateway returns the default gateway address
func (c *Config) Gateway() net.IP {
	return net.ParseIP(c.viper.GetString("gateway"))
}

// DNSServers returns the dns server addresses
func (c *Config) DNSServers() []net.IP {
	addresses := c.viper.GetStringSlice("dns-servers")
	ips := make([]net.IP, len(addresses))

	for i := range addresses {
		ip := net.ParseIP(addresses[i])
		if ip == nil {
			return nil
		}
		ips[i] = ip
	}

	return ips
}

// Network returns the ip range
func (c *Config) Network() *net.IPNet {
	net, _ := c.network()
	return net
}

func (c *Config) network() (*net.IPNet, error) {
	_, net, err := net.ParseCIDR(c.viper.GetString("network"))
	if err != nil {
		return nil, err
	}

	return net, nil
}
