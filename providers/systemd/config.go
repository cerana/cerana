package systemd

import (
	"errors"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify/provider"
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
	UnitFileDir string `json:"unit_file_dir"`
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	return &Config{provider.NewConfig(flagSet, v)}
}

// UnitFilePath returns the absolute path of the unit file for a unit.
func (c *Config) UnitFilePath(name string) (string, error) {
	var unitFileDir string
	if err := c.UnmarshalKey("unit_file_dir", &unitFileDir); err != nil {
		return "", err
	}

	baseName := filepath.Base(name)
	if baseName != name || name == "/" || name == "." || name == ".." {
		return "", errors.New("invalid name")
	}

	return filepath.Abs(filepath.Join(unitFileDir, name))
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	var unitFileDir string
	if err := c.UnmarshalKey("unit_file_dir", &unitFileDir); err != nil {
		err = errors.New("invalid unit_file_dir")
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("invalid config")
		return err
	}
	if unitFileDir == "" {
		err := errors.New("missing unit_file_dir")
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("invalid config")
		return err
	}

	return nil
}

// LoadConfig loads and validates the Systemd provider config
func (c *Config) LoadConfig() error {
	if err := c.Config.LoadConfig(); err != nil {
		return err
	}

	return c.Validate()
}
