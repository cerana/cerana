package tick

import (
	"net/url"
	"os"
	"time"

	"github.com/cerana/cerana/pkg/configutil"
	"github.com/cerana/cerana/pkg/errors"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Configer is an interface that provides information the tick needs to run. It
// allows more complex configs to be passed through to the tick function.
type Configer interface {
	ConfigFile() string
	NodeDataURL() *url.URL
	ClusterDataURL() *url.URL
	LogLevel() string
	RequestTimeout() time.Duration
	TickInterval() time.Duration
	TickRetryInterval() time.Duration
}

// Config is the configuration for a tick.
type Config struct {
	viper   *viper.Viper
	flagSet *pflag.FlagSet
}

// ConfigData defines the structure of the config data (e.g. in the config file)
type ConfigData struct {
	NodeDataURL       string `json:"nodeDataURL"`
	ClusterDataURL    string `json:"clusterDataURL"`
	LogLevel          string `json:"logLevel"`
	RequestTimeout    string `json:"requestTimeout"`
	TickInterval      string `json:"tickInterval"`
	TickRetryInterval string `json:"tickRetryInterval"`
}

// NewConfig creates a new instance of Config.
func NewConfig(flagSet *pflag.FlagSet, v *viper.Viper) *Config {
	if flagSet == nil {
		flagSet = pflag.CommandLine
	}

	if v == nil {
		v = viper.New()
	}

	// Set normalization function before adding any flags
	configutil.Normalize(flagSet)

	flagSet.StringP("configFile", "c", "", "path to config file")
	flagSet.StringP("nodeDataURL", "n", "", "url of coordinator for node information retrieval")
	flagSet.StringP("clusterDataURL", "u", "", "url of coordinator for the cluster information")
	flagSet.StringP("logLevel", "l", "warning", "log level: debug/info/warn/error/fatal/panic")
	flagSet.DurationP("requestTimeout", "r", 0, "default timeout for external requests made")
	flagSet.DurationP("tickInterval", "t", 0, "tick run frequency")
	flagSet.DurationP("tickRetryInterval", "i", 0, "tick retry on error frequency")

	return &Config{
		viper:   v,
		flagSet: flagSet,
	}
}

// LoadConfig loads the config.
func (c *Config) LoadConfig() error {
	if !c.flagSet.Parsed() {
		if err := c.flagSet.Parse(os.Args[1:]); err != nil {
			return errors.Wrap(err)
		}
	}
	if err := c.viper.BindPFlags(c.flagSet); err != nil {
		return errors.Wrap(err, "failed to bind flags")
	}

	filePath := c.ConfigFile()
	if filePath == "" {
		return c.Validate()
	}

	c.viper.SetConfigFile(filePath)
	if err := c.viper.ReadInConfig(); err != nil {
		return errors.Wrapv(err, map[string]interface{}{"path": filePath}, "failed to parse config file")
	}

	return c.Validate()
}

// ConfigFile returns a path to a config file.
func (c *Config) ConfigFile() string {
	return c.viper.GetString("configFile")
}

// NodeDataURL returns the url of the layer 1 coordinator, used for node
// information.
func (c *Config) NodeDataURL() *url.URL {
	// Error checking has been done during validation
	u, _ := url.ParseRequestURI(c.viper.GetString("nodeDataURL"))
	return u
}

// ClusterDataURL returns the url of the layer 2 coordinator, used for cluster
// information.
func (c *Config) ClusterDataURL() *url.URL {
	// Error checking has been done during validation
	u, _ := url.ParseRequestURI(c.viper.GetString("clusterDataURL"))
	return u
}

// RequestTimeout returns the default timeout for task requests.
func (c *Config) RequestTimeout() time.Duration {
	return c.viper.GetDuration("requestTimeout")
}

// TickInterval returns how often the tick function should be executed.
func (c *Config) TickInterval() time.Duration {
	return c.viper.GetDuration("tickInterval")
}

// TickRetryInterval returns how often the tick function should be executed
// after an error.
func (c *Config) TickRetryInterval() time.Duration {
	return c.viper.GetDuration("tickRetryInterval")
}

// LogLevel returns the log level.
func (c *Config) LogLevel() string {
	return c.viper.GetString("logLevel")
}

// SetupLogging sets up logging with the log level and formatting.
func (c *Config) SetupLogging() error {
	return logrusx.SetLevel(c.LogLevel())
}

// Validate ensures the configuration is valid.
func (c *Config) Validate() error {
	if c.TickInterval() == 0 {
		return errors.New("tickInterval must be greater than 0")
	}
	if c.TickRetryInterval() == 0 {
		return errors.New("tickRetryInterval must be greater than 0")
	}
	if c.RequestTimeout() == 0 {
		return errors.New("requestTimeout must be greater than 0")
	}

	if err := c.ValidateURL("nodeDataURL"); err != nil {
		return err
	}
	if err := c.ValidateURL("clusterDataURL"); err != nil {
		return err
	}

	return nil
}

// ValidateURL is used in validation for checking url parameters.
func (c *Config) ValidateURL(name string) error {
	u := c.viper.GetString(name)
	if u == "" {
		return errors.New("missing " + name)
	}
	if _, err := url.ParseRequestURI(u); err != nil {
		return errors.Wrapv(err, map[string]interface{}{"url": u}, "invalid "+name)
	}
	return nil
}
