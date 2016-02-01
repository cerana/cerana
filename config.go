package provider

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	logx "github.com/mistifyio/mistify-logrus-ext"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config holds all configuration for the provider.
type Config struct {
	viper   *viper.Viper
	flagSet *flag.FlagSet
}

// NewConfig creates a new instance of Config. If a viper instance is not
// provided, a new one will be created.
func NewConfig(flagSet *flag.FlagSet, v *viper.Viper) *Config {
	if flagSet == nil {
		flagSet = flag.CommandLine
	}

	if v == nil {
		v = viper.New()
	}

	flagSet.StringP("config_file", "c", "", "path to config file")
	flagSet.StringP("socket_dir", "s", "/tmp/mistify", "base directory in which to create task sockets")
	flagSet.UintP("default_priority", "p", 50, "default task priority")
	flagSet.StringP("log_level", "l", "warning", "log level: debug/info/warn/error/fatal/panic")
	flagSet.Uint64P("request_timeout", "t", 0, "default timeout for requests made by this provider in seconds")

	return &Config{
		viper:   viper.New(),
		flagSet: flagSet,
	}
}

// LoadConfig attempts to load the config. Flags should be parsed first.
func (c *Config) LoadConfig() error {
	if err := c.viper.BindPFlags(c.flagSet); err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("failed to bind flags")
		return err
	}

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

// TaskTimeout determines the timeout for a task. If a timeout was not
// explicitly configured for the task, it will return the default.
func (c *Config) TaskTimeout(taskName string) time.Duration {
	key := fmt.Sprintf("tasks.%s.timeout", taskName)
	var seconds int
	if c.viper.IsSet(key) {
		seconds = c.viper.GetInt(key)
	} else {
		seconds = c.viper.GetInt("default_timeout")
	}

	return time.Duration(seconds) * time.Second
}

// SocketDir returns the base directory for task sockets.
func (c *Config) SocketDir() string {
	return c.viper.GetString("socket_dir")
}

// StreamDir returns the directory for ad-hoc data stream sockets.
func (c *Config) StreamDir() string {
	return filepath.Join(
		c.SocketDir(),
		"streams",
		"StreamEcho",
		c.ServiceName())
}

// ServiceName returns the name the service should register as.
func (c *Config) ServiceName() string {
	return c.viper.GetString("service_name")
}

// CoordinatorURL returns the URL of the Coordinator for which the Provider is
// registered.
func (c *Config) CoordinatorURL() *url.URL {
	// Error checking has been done during validation
	u, _ := url.ParseRequestURI(c.viper.GetString("coordinator_url"))
	return u
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
	if _, err := url.ParseRequestURI(c.viper.GetString("coordinator_url")); err != nil {
		log.WithFields(log.Fields{
			"coordinator_url": c.viper.GetString("coordinator_url"),
			"error":           err,
		}).Error("invalid config")
		return err
	}

	return nil
}

// Unmarshal unmarshals the config into a struct.
func (c *Config) Unmarshal(rawVal interface{}) error {
	return c.viper.Unmarshal(rawVal)
}

// UnmarshalKey unmarshals a single config key into a struct.
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	return c.viper.UnmarshalKey(key, rawVal)
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
