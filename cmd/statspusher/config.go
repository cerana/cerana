package main

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	flagSep     = regexp.MustCompile(`[\s._-]+`)
	specialCaps = regexp.MustCompile("(?i)^(url|ttl|cpu|ip|id)$")
)

type config struct {
	viper   *viper.Viper
	flagSet *pflag.FlagSet
}

// ConfigData defines the structure of the config data (e.g. in the config file)
type ConfigData struct {
	CoordinatorURL string `json:"coordinatorURL"`
	HeartbeatURL   string `json:"heartbeatURL"`
	LogLevel       string `json:"logLevel"`
	// Timeout and TTL values are in seconds
	RequestTimeout uint `json:"requestTimeout"`
	DatasetTTL     uint `json:"datasetTTL"`
	BundleTTL      uint `json:"bundleTTL"`
	NodeTTL        uint `json:"nodeTTL"`
}

func newConfig(flagSet *pflag.FlagSet, v *viper.Viper) *config {
	if flagSet == nil {
		flagSet = pflag.CommandLine
	}

	if v == nil {
		v = viper.New()
	}

	// Set normalization function before adding any flags
	flagSet.SetNormalizeFunc(canonicalFlagName)

	// Update Usage (--help output) to indicate flag
	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		pflag.PrintDefaults()
		fmt.Fprintln(os.Stderr, "Note: Flags can be used in either fooBar or foo[_-.]bar form.")
	}

	flagSet.StringP("configFile", "c", "", "path to config file")
	flagSet.StringP("coordinatorURL", "u", "", "url of coordinator for information retrieval")
	flagSet.StringP("heartbeatURL", "e", "", "url of coordinator for the heartbeat registering")
	flagSet.StringP("logLevel", "l", "warning", "log level: debug/info/warn/error/fatal/panic")
	flagSet.Uint64P("requestTimeout", "r", 0, "default timeout for requests made (seconds)")
	flagSet.Uint64P("datasetTTL", "d", 0, "default timeout for requests made (seconds)")
	flagSet.Uint64P("bundleTTL", "b", 0, "default timeout for requests made (seconds)")
	flagSet.Uint64P("nodeTTL", "n", 0, "default timeout for requests made  (seconds)")

	return &config{
		viper:   v,
		flagSet: flagSet,
	}
}

// canonicalFlagName translates flag names to camelCase using whitespace,
// periods, underscores, and dashes as word boundaries. All-caps words are
// preserved.
func canonicalFlagName(f *pflag.FlagSet, name string) pflag.NormalizedName {
	// Standardize separators to a single space and trim leading/trailing spaces
	name = strings.TrimSpace(flagSep.ReplaceAllString(name, " "))

	// Convert to title case (lower case with leading caps, preserved all caps)
	name = strings.Title(name)

	// Some words should always be all caps or all lower case (e.g. TTL)
	nameParts := strings.Split(name, " ")
	for i, part := range nameParts {
		caseFn := strings.ToUpper
		if i == 0 {
			caseFn = strings.ToLower
		}

		nameParts[i] = specialCaps.ReplaceAllStringFunc(part, caseFn)
	}

	// Split on space and examine the first part
	first := nameParts[0]
	if utf8.RuneCountInString(first) == 1 || first != strings.ToUpper(first) {
		// Lowercase the first letter if it is not an all-caps word
		r, n := utf8.DecodeRuneInString(first)
		nameParts[0] = string(unicode.ToLower(r)) + first[n:]
	}

	return pflag.NormalizedName(strings.Join(nameParts, ""))
}

func (c *config) loadConfig() error {
	if err := c.viper.BindPFlags(c.flagSet); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Error("failed to bind flags")
		return err
	}

	filePath := c.viper.GetString("configFile")
	if filePath == "" {
		return c.validate()
	}

	c.viper.SetConfigFile(filePath)
	if err := c.viper.ReadInConfig(); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":    err,
			"filePath": filePath,
		}).Error("failed to parse config file")
		return err
	}

	return c.validate()
}

func (c *config) coordinatorURL() *url.URL {
	// Error checking has been done during validation
	u, _ := url.ParseRequestURI(c.viper.GetString("coordinatorURL"))
	return u
}

func (c *config) heartbeatURL() *url.URL {
	// Error checking has been done during validation
	u, _ := url.ParseRequestURI(c.viper.GetString("heartbeatURL"))
	return u
}

func (c *config) requestTimeout() time.Duration {
	return time.Second * time.Duration(c.viper.GetInt("requestTimeout"))
}

func (c *config) datasetTTL() time.Duration {
	return c.getTTL("datasetTTL")
}

func (c *config) nodeTTL() time.Duration {
	return c.getTTL("nodeTTL")
}

func (c *config) bundleTTL() time.Duration {
	return c.getTTL("bundleTTL")
}

func (c *config) getTTL(key string) time.Duration {
	return time.Second * time.Duration(c.viper.GetInt(key))
}

func (c *config) setupLogging() error {
	logLevel := c.viper.GetString("logLevel")
	if err := logrusx.SetLevel(logLevel); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
			"level": logLevel,
		}).Error("failed to set up logging")
		return err
	}
	return nil
}

func (c *config) validate() error {
	if c.datasetTTL() <= 0 {
		return errors.New("dataset ttl must be > 0")
	}
	if c.bundleTTL() <= 0 {
		return errors.New("bundle ttl must be > 0")
	}
	if c.nodeTTL() <= 0 {
		return errors.New("node ttl must be > 0")
	}
	if c.requestTimeout() <= 0 {
		return errors.New("request timeout must be > 0")
	}

	if err := c.validateURL("coordinatorURL"); err != nil {
		return err
	}
	if err := c.validateURL("heartbeatURL"); err != nil {
		return err
	}

	return nil
}

func (c *config) validateURL(name string) error {
	u := c.viper.GetString(name)
	if u == "" {
		return errors.New("missing " + name)
	}
	if _, err := url.ParseRequestURI(u); err != nil {
		logrus.WithFields(logrus.Fields{
			name:    u,
			"error": err,
		}).Error("invalid config")
		return errors.New("invalid " + name)
	}
	return nil
}
