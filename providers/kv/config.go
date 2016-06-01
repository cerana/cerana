package kv

import (
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
	return c.Config.Validate()
}

// LoadConfig loads and validates the KV provider config.
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}

// Address returns the configured address of the consul kv server.
func (c *Config) Address() (string, error) {
	var address string
	if err := c.UnmarshalKey("address", &address); err != nil {
		return "", err
	}

	return address, nil
}
