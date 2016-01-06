package simple

import (
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/spf13/viper"
)

// Config holds all configuration for the provider.
type Config struct {
	viper *viper.Viper
}

// NewConfig creates a new instance of Config. If a viper instance is not
// provided, a new one will be created.
func NewConfig(v *viper.Viper) *Config {
	if v == nil {
		v = viper.New()
	}
	return &Config{
		viper: v,
	}
}

// LoadConfigFile attempts to load a config file.
func (c *Config) LoadConfigFile() error {
	filePath := c.viper.GetString("config_file")
	if filePath == "" {
		return c.Validate()
	}

	c.viper.SetConfigFile(filePath)
	if err := c.viper.ReadInConfig(); err != nil {
		log.WithFields(log.Fields{
			"error":    err,
			"filePath": filePath,
		}).Error("failed to parse config file")
		return err
	}

	return c.Validate()
}

// TaskPriority determines the registration priority of a task. If a
// priority was not explicitly configured for the task, it will return the
// default.
func (c *Config) TaskPriority(taskName string) int {
	key := fmt.Sprintf("tasks.%s.priority", taskName)
	if c.viper.IsSet(key) {
		return c.viper.GetInt(key)
	}
	return c.viper.GetInt("default_priority")
}

// SocketDir returns the base directory for task sockets.
func (c *Config) SocketDir() string {
	return c.viper.GetString("socket_dir")
}

// Validate returns whether the config is valid, containing necessary values.
func (c *Config) Validate() error {
	if c.SocketDir() == "" {
		err := errors.New("missing socket_dir")
		log.WithFields(log.Fields{
			"error": err,
		}).Error("invalid config")
		return err
	}

	return nil
}

// SetupLogging sets the log level and formatting.
func (c *Config) SetupLogging() error {
	logLevel := c.viper.GetString("log_level")
	if err := logx.SetLevel(logLevel); err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"level": logLevel,
		}).Error("failed to set up logging")
		return err
	}
	return nil
}
