package coordinator

import (
	"errors"
	"time"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	"github.com/spf13/viper"
)

// Config holds all configuration for the provider.
type Config struct {
	viper *viper.Viper
}

// ConfigData defines the structure of the config data (e.g. in the config file)
type ConfigData struct {
	SocketDir      string `json:"socket_dir"`
	ServiceName    string `json:"service_name"`
	ExternalPort   int    `json:"external_port"`
	RequestTimeout uint64 `json:"request_timeout"`
	LogLevel       string `json:"log_level"`
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

// SocketDir returns the base directory for task sockets.
func (c *Config) SocketDir() string {
	return c.viper.GetString("socket_dir")
}

// ServiceName returns the name the service should register as.
func (c *Config) ServiceName() string {
	return c.viper.GetString("service_name")
}

// ExternalPort returns the port to listen on for external requests.
func (c *Config) ExternalPort() int {
	return c.viper.GetInt("external_port")
}

// RequestTimeout returns the duration of the default request timeout.
func (c *Config) RequestTimeout() time.Duration {
	return time.Second * time.Duration(c.viper.GetInt("request_timeout"))
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

	if c.ServiceName() == "" {
		err := errors.New("missing service_name")
		log.WithFields(log.Fields{
			"error": err,
		}).Error("invalid config")
		return err
	}

	if c.ExternalPort() == 0 {
		err := errors.New("missing external_port")
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
