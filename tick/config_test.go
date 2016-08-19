package tick_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/tick"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (s *Tick) TestNewConfig() {
	tests := []struct {
		description string
		flagSet     *pflag.FlagSet
		viper       *viper.Viper
	}{
		{"defaults", nil, nil},
		{"specified flagset", pflag.NewFlagSet(uuid.New(), pflag.ExitOnError), nil},
		{"specified viper", nil, viper.New()},
		{"specified both", pflag.NewFlagSet(uuid.New(), pflag.ExitOnError), viper.New()},
	}

	for _, test := range tests {
		// Reset the default CommandLine flags between runs
		pflag.CommandLine = pflag.NewFlagSet(uuid.New(), pflag.ExitOnError)
		s.NotNil(tick.NewConfig(test.flagSet, test.viper), test.description)
	}
}

func (s *Tick) TestLoadConfig() {
	tests := []struct {
		description string
		setFlags    bool
		writeConfig bool
		expectedErr bool
	}{
		{"nothing set", false, false, true},
		{"flags only", true, false, false},
		{"config file", false, true, false},
		{"flags and config", true, true, false},
	}

	for _, test := range tests {
		config, _, _, configFile, err := newTestConfig(test.setFlags, test.writeConfig, s.configData)
		if configFile != nil {
			defer func() { _ = os.Remove(configFile.Name()) }()
		}
		if !s.NoError(err, test.description) {
			continue
		}

		err = config.LoadConfig()
		if test.expectedErr {
			s.Error(err, test.description)
		} else {
			s.NoError(err, test.description)
		}
	}
}

func (s *Tick) TestSetupLogging() {
	s.NoError(s.config.SetupLogging(), "failed to setup logging")
	s.Equal(s.configData.LogLevel, logrus.GetLevel().String())
}

func (s *Tick) TestValidate() {
	u := "unix:///tmp/foobar"
	tests := []struct {
		description       string
		nodeDataURL       string
		clusterDataURL    string
		requestTimeout    string
		tickInterval      string
		tickRetryInterval string
		expectedErr       string
	}{
		{"valid", u, u, "5s", "4s", "3s", ""},
		{"missing nodeDataURL", "", u, "5s", "4s", "3s", "missing nodeDataURL"},
		{"invalud nodeDataURL", "asdf", u, "5s", "4s", "3s", "invalid nodeDataURL"},
		{"missing clusterDataURL", u, "", "5s", "4s", "3s", "missing clusterDataURL"},
		{"invalud clusterDataURL", u, "asdf", "5s", "4s", "3s", "invalid clusterDataURL"},
		{"invalid request timeout", u, u, "0", "4s", "3s", "requestTimeout must be greater than 0"},
		{"invalid tick interval", u, u, "5s", "0", "3s", "tickInterval must be greater than 0"},
		{"missing tick retry interval", u, u, "5s", "4s", "0", "tickRetryInterval must be greater than 0"},
	}

	for _, test := range tests {
		configData := &tick.ConfigData{
			NodeDataURL:       test.nodeDataURL,
			ClusterDataURL:    test.clusterDataURL,
			RequestTimeout:    test.requestTimeout,
			TickInterval:      test.tickInterval,
			TickRetryInterval: test.tickRetryInterval,
		}

		config, fs, v, _, err := newTestConfig(true, false, configData)
		if !s.NoError(err, test.description) {
			continue
		}
		// Bind here to avoid the need for Load
		s.Require().NoError(v.BindPFlags(fs), test.description)

		err = config.Validate()
		if test.expectedErr != "" {
			if s.NotNil(err, test.description) {
				s.Contains(err.Error(), test.expectedErr, test.description)
			}
		} else {
			s.NoError(err, test.description)
		}
	}
}

func (s *Tick) TestNodeDataURL() {
	u, err := url.ParseRequestURI(s.configData.NodeDataURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.NodeDataURL())
}

func (s *Tick) TestClusterDataURL() {
	u, err := url.ParseRequestURI(s.configData.ClusterDataURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.ClusterDataURL())
}

func (s *Tick) TestRequestTimeout() {
	s.EqualValues(s.configData.RequestTimeout, s.config.RequestTimeout().String())
}

func (s *Tick) TestTickInterval() {
	s.EqualValues(s.configData.TickInterval, s.config.TickInterval().String())
}

func newTestConfig(setFlags, writeConfig bool, configData *tick.ConfigData) (*tick.Config, *pflag.FlagSet, *viper.Viper, *os.File, error) {
	fs := pflag.NewFlagSet(uuid.New(), pflag.ExitOnError)
	v := viper.New()
	v.SetConfigType("json")
	config := tick.NewConfig(fs, v)
	if config == nil {
		return nil, nil, nil, nil, errors.New("failed to return a config")
	}

	var configFile *os.File
	if writeConfig {
		var err error
		configFile, err = ioutil.TempFile("", "statspusherTest-")
		if err != nil {
			return nil, nil, nil, nil, err
		}
		defer func() { _ = configFile.Close() }()

		configJSON, _ := json.Marshal(configData)
		if _, err := configFile.Write(configJSON); err != nil {
			return nil, nil, nil, configFile, err
		}

		if err := fs.Set("configFile", configFile.Name()); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	if err := fs.Parse([]string{}); err != nil {
		return nil, nil, nil, nil, err
	}

	if setFlags {
		if err := fs.Set("nodeDataURL", configData.NodeDataURL); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("clusterDataURL", configData.ClusterDataURL); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("logLevel", configData.LogLevel); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("requestTimeout", configData.RequestTimeout); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("tickInterval", configData.TickInterval); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("tickRetryInterval", configData.TickRetryInterval); err != nil {
			return nil, nil, nil, configFile, err
		}

	}

	return config, fs, v, configFile, nil
}
