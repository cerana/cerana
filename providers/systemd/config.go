package systemd

import (
	"errors"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/mistifyio/mistify/provider"
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

// UnitFilePath returns the absolute path of the unit file for a unit.
func (c *Config) UnitFilePath(name string) (string, error) {
	var unitFileDir string
	if err := c.UnmarshalKey("unit_file_dir", &unitFileDir); err != nil {
		return "", err
	}
	name = filepath.Base(name)
	if name == "." || name == ".." {
		return "", errors.New("invalid filename")
	}
	return filepath.Abs(filepath.Join(unitFileDir, name))
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if err := c.Config.Validate(); err != nil {
		return err
	}

	if path, err := c.UnitFilePath(""); err != nil || path == "/" {
		err = errors.New("missing or invalid unit_file_dir")
		logrus.WithFields(logrus.Fields{
			"unit_file_dir": path,
			"error":         err,
		}).Error("invalid config")
		return err
	}

	return nil
}
