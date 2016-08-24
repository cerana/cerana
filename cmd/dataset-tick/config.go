package main

import (
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/tick"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config contains configuration required for the dataset tick.
type Config struct {
	*tick.Config
	flagSet *pflag.FlagSet
	viper   *viper.Viper
}

// ConfigData defines the structure of the config data (e.g. in the config file).
type ConfigData struct {
	tick.ConfigData
	DatasetPrefix string `json:"datasetPrefix"`
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	if flagSet == nil {
		flagSet = pflag.CommandLine
	}

	if v == nil {
		v = viper.New()
	}

	config := &Config{
		Config:  tick.NewConfig(flagSet, v),
		flagSet: flagSet,
		viper:   v,
	}
	config.flagSet.StringP("datasetPrefix", "d", "", "dataset directory")

	return config
}

// LoadConfig loads and validates the config.
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}

// DatasetPrefix returns the prefix for monitored datasets.
func (c *Config) DatasetPrefix() string {
	return c.viper.GetString("datasetPrefix")
}

// Validate ensures the configuration is valid.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}
	if c.DatasetPrefix() == "" {
		return errors.New("missing datasetPrefix")
	}
	if c.HTTPResponseURL() == nil {
		return errors.New("missing responseAddr")
	}

	return nil
}
