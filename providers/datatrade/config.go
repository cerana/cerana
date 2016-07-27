package datatrade

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
	DatasetDir          string `json:"dataset_dir"`
	NodeCoordinatorPort uint   `json:"node_coordinator_port"`
}

// DatasetDir returns the directory in which datasets are stored on nodes.
func (c *Config) DatasetDir() string {
	var datasetDir string
	_ = c.UnmarshalKey("dataset_dir", &datasetDir)
	// Checked at validation time.
	return datasetDir
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

// NodeCoordinatorPort returns the port that node coordinators are running on.
func (c *Config) NodeCoordinatorPort() uint {
	var port uint
	_ = c.UnmarshalKey("node_coordinator_port", &port)
	return port
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if c.DatasetDir() == "" {
		return errors.New("missing dataset_dir")
	}

	if c.NodeCoordinatorPort() == 0 {
		return errors.New("missing node_coordinator_port")
	}

	return nil
}
