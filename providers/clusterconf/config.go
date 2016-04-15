package clusterconf

import (
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
