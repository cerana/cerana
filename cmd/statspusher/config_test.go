package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (s *StatsPusher) TestCanonicalFlagName() {
	tests := []struct {
		input  string
		output string
	}{
		{"foo bar", "fooBar"},
		{"foo	bar", "fooBar"},
		{"foo.bar", "fooBar"},
		{"foo_bar", "fooBar"},
		{"foo-bar", "fooBar"},
		{"foo.-_bar", "fooBar"},
		{"Foo_bar", "fooBar"},
		{"foo_bar.baz bang", "fooBarBazBang"},
		{"Foo_bar", "fooBar"},
		{"foo_BAR", "fooBAR"},
		{"foo_cpu", "fooCPU"},
		{"foo_ttl", "fooTTL"},
		{"foo_cpu", "fooCPU"},
		{"foo_url", "fooURL"},
		{"foo_ip", "fooIP"},
		{"foo_id", "fooID"},
		{"FOO_bar", "FOOBar"},
		{"Cpu_bar", "cpuBar"},
		{"id_bar", "idBar"},
		{"CPU_bar", "cpuBar"},
	}

	for _, test := range tests {
		s.Equal(test.output, string(canonicalFlagName(pflag.CommandLine, test.input)), test.input)
	}
}

func (s *StatsPusher) TestNewConfig() {
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
		s.NotNil(newConfig(test.flagSet, test.viper), test.description)
	}
}

func (s *StatsPusher) TestLoadConfig() {
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

		err = config.loadConfig()
		if test.expectedErr {
			s.Error(err, test.description)
		} else {
			s.NoError(err, test.description)
		}
	}
}

func (s *StatsPusher) TestSetupLogging() {
	s.NoError(s.config.setupLogging(), "failed to setup logging")
	s.Equal(s.configData.LogLevel, logrus.GetLevel().String())
}

func (s *StatsPusher) TestValidate() {
	u := "unix:///tmp/foobar"
	tests := []struct {
		description    string
		coordinatorURL string
		heartbeatURL   string
		requestTimeout uint
		datasetTTL     uint
		bundleTTL      uint
		nodeTTL        uint
		expectedErr    string
	}{
		{"valid", u, u, 5, 4, 3, 2, ""},
		{"missing coord", "", u, 5, 4, 3, 2, "missing coordinatorURL"},
		{"invalud coord", "asdf", u, 5, 4, 3, 2, "invalid coordinatorURL"},
		{"missing heartbeat", u, "", 5, 4, 3, 2, "missing heartbeatURL"},
		{"invalud heartbeat", u, "asdf", 5, 4, 3, 2, "invalid heartbeatURL"},
		{"invalid request timeout", u, u, 0, 4, 3, 2, "request timeout must be > 0"},
		{"invalid dataset ttl", u, u, 5, 0, 3, 2, "dataset ttl must be > 0"},
		{"invalid bundle ttl", u, u, 5, 4, 0, 2, "bundle ttl must be > 0"},
		{"invalid node ttl", u, u, 5, 4, 3, 0, "node ttl must be > 0"},
	}

	for _, test := range tests {
		configData := &ConfigData{
			CoordinatorURL: test.coordinatorURL,
			HeartbeatURL:   test.heartbeatURL,
			RequestTimeout: test.requestTimeout,
			DatasetTTL:     test.datasetTTL,
			BundleTTL:      test.bundleTTL,
			NodeTTL:        test.nodeTTL,
		}

		config, fs, v, _, err := newTestConfig(true, false, configData)
		if !s.NoError(err, test.description) {
			continue
		}
		// Bind here to avoid the need for Load
		s.Require().NoError(v.BindPFlags(fs), test.description)

		err = config.validate()
		if test.expectedErr != "" {
			s.EqualError(err, test.expectedErr, test.description)
		} else {
			s.NoError(err, test.description)
		}
	}
}

func (s *StatsPusher) TestCoordinatorURL() {
	u, err := url.ParseRequestURI(s.configData.CoordinatorURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.coordinatorURL())
}

func (s *StatsPusher) TestHeartbeatURL() {
	u, err := url.ParseRequestURI(s.configData.HeartbeatURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.heartbeatURL())
}

func (s *StatsPusher) TestRequestTimeout() {
	s.EqualValues(s.configData.RequestTimeout, s.config.requestTimeout()/time.Second)
}

func (s *StatsPusher) TestDatasetTTL() {
	s.EqualValues(s.configData.DatasetTTL, s.config.datasetTTL()/time.Second)
}

func (s *StatsPusher) TestBundleTTL() {
	s.EqualValues(s.configData.BundleTTL, s.config.bundleTTL()/time.Second)
}

func (s *StatsPusher) TestNodeTTL() {
	s.EqualValues(s.configData.NodeTTL, s.config.nodeTTL()/time.Second)
}

func newTestConfig(setFlags, writeConfig bool, configData *ConfigData) (*config, *pflag.FlagSet, *viper.Viper, *os.File, error) {
	fs := pflag.NewFlagSet(uuid.New(), pflag.ExitOnError)
	v := viper.New()
	v.SetConfigType("json")
	config := newConfig(fs, v)
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

	if setFlags {
		if err := fs.Set("coordinatorURL", configData.CoordinatorURL); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("heartbeatURL", configData.HeartbeatURL); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("logLevel", configData.LogLevel); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("requestTimeout", strconv.FormatUint(uint64(configData.RequestTimeout), 10)); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("datasetTTL", strconv.FormatUint(uint64(configData.DatasetTTL), 10)); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("bundleTTL", strconv.FormatUint(uint64(configData.BundleTTL), 10)); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("nodeTTL", strconv.FormatUint(uint64(configData.NodeTTL), 10)); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	return config, fs, v, configFile, nil
}
