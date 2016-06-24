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
	DatasetTTL string `json:"datasetTTL"`
	BundleTTL  string `json:"bundleTTL"`
	NodeTTL    string `json:"nodeTTL"`
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{provider.NewConfig(flagSet, v)}
}

// DatasetTTL returns the TTL for dataset node heartbeats.
func (c *Config) DatasetTTL() time.Duration {
	var ttlString string
	_ = c.UnmarshalKey("dataset_ttl", &ttlString)
	// Since errors lead to a 0 value and 0 is considered invalid, safe to
	// ignore the error.
	ttl, _ := time.ParseDuration(ttlString)
	return ttl
}

// BundleTTL returns the TTL for bundle node heartbeats.
func (c *Config) BundleTTL() time.Duration {
	var ttlString string
	_ = c.UnmarshalKey("bundle_ttl", &ttlString)
	// Since errors lead to a 0 value and 0 is considered invalid, safe to
	// ignore the error.
	ttl, _ := time.ParseDuration(ttlString)
	return ttl
}

// NodeTTL returns the TTL for node heartbeats.
func (c *Config) NodeTTL() time.Duration {
	var ttlString string
	_ = c.UnmarshalKey("node_ttl", &ttlString)
	// Since errors lead to a 0 value and 0 is considered invalid, safe to
	// ignore the error.
	ttl, _ := time.ParseDuration(ttlString)
	return ttl
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

// LoadConfig loads and validates the ClusterConf provider config.
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}
