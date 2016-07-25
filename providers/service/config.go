package service

import (
	"errors"

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
	DatasetCloneBin  string `json:"dataset_clone_bin"`
	DatasetClonePath string `json:"dataset_clone_path"`
}

// DatasetCloneBin returns the full path of the binary to clone/rollback
// datasets for services.
func (c *Config) DatasetCloneBin() string {
	var dcb string
	_ = c.UnmarshalKey("dataset_clone_bin", &dcb)
	// Checked at validation time
	return dcb
}

// DatasetClonePath returns the zfs path in which to clone datasets.
func (c *Config) DatasetClonePath() string {
	var dcp string
	_ = c.UnmarshalKey("dataset_clone_path", &dcp)
	// Checked at validation time
	return dcp
}

// LoadConfig loads and validates the config data.
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{provider.NewConfig(flagSet, v)}
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.DatasetCloneBin() == "" {
		return errors.New("missing dataset_clone_bin")
	}

	if c.DatasetClonePath() == "" {
		return errors.New("missing dataset_clone_path")
	}

	return nil
}
