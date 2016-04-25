package clusterconf

import (
	"errors"
	"time"

	"github.com/cerana/cerana/provider"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the provider.
type Config struct {
	*provider.Config
}

// ConfigData defines the structure of the config data (e.g. in the config file)
type ConfigData struct {
	provider.ConfigData
	DatasetTTL time.Duration
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{provider.NewConfig(flagSet, v)}
}

// DatasetTTL returns the TTL for dataset node heartbeats.
func (c *Config) DatasetTTL() time.Duration {
	var datasetTTL time.Duration
	_ = c.UnmarshalKey("dataset_ttl", &datasetTTL)
	return datasetTTL
}

// BundleTTL returns the TTL for bundle node heartbeats.
func (c *Config) BundleTTL() time.Duration {
	var datasetTTL time.Duration
	_ = c.UnmarshalKey("bundle_ttl", &datasetTTL)
	return datasetTTL
}

// NodeTTL returns the TTL for node heartbeats.
func (c *Config) NodeTTL() time.Duration {
	var nodeTTL time.Duration
	_ = c.UnmarshalKey("node_ttl", &nodeTTL)
	return nodeTTL
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.DatasetTTL() <= 0 {
		return errors.New("invalid dataset_ttl")
	}
	if c.BundleTTL() <= 0 {
		return errors.New("invalid bundle_ttl")
	}
	if c.NodeTTL() <= 0 {
		return errors.New("invalid node_ttl")
	}

	return nil
}
