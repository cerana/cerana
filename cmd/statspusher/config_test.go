package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/url"
	"os"

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
		{"foo_interval", "fooInterval"},
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
		description     string
		nodeDataURL     string
		clusterDataURL  string
		requestTimeout  string
		datasetInterval string
		bundleInterval  string
		nodeInterval    string
		datasetDir      string
		expectedErr     string
	}{
		{"valid", u, u, "5s", "4s", "3s", "2s", "foobar", ""},
		{"missing nodeDataURL", "", u, "5s", "4s", "3s", "2s", "foobar", "missing nodeDataURL"},
		{"invalud nodeDataURL", "asdf", u, "5s", "4s", "3s", "2s", "foobar", "invalid nodeDataURL"},
		{"missing clusterDataURL", u, "", "5s", "4s", "3s", "2s", "foobar", "missing clusterDataURL"},
		{"invalud clusterDataURL", u, "asdf", "5s", "4s", "3s", "2s", "foobar", "invalid clusterDataURL"},
		{"invalid request timeout", u, u, "0", "4s", "3s", "2s", "foobar", "request timeout must be > 0"},
		{"invalid dataset interval", u, u, "5s", "0", "3s", "2s", "foobar", "dataset interval must be > 0"},
		{"invalid bundle interval", u, u, "5s", "4s", "0", "2s", "foobar", "bundle interval must be > 0"},
		{"invalid node interval", u, u, "5s", "4s", "3s", "0", "foobar", "node interval must be > 0"},
		{"invalid node interval", u, u, "5s", "4s", "3s", "2s", "", "missing datasetDir"},
		{"invalid node interval", u, u, "5s", "4s", "3s", "2s", "foobar", ""},
	}

	for _, test := range tests {
		configData := &ConfigData{
			NodeDataURL:     test.nodeDataURL,
			ClusterDataURL:  test.clusterDataURL,
			RequestTimeout:  test.requestTimeout,
			DatasetInterval: test.datasetInterval,
			BundleInterval:  test.bundleInterval,
			NodeInterval:    test.nodeInterval,
			DatasetDir:      test.datasetDir,
		}

		config, fs, v, _, err := newTestConfig(true, false, configData)
		if !s.NoError(err, test.description) {
			continue
		}
		// Bind here to avoid the need for Load
		s.Require().NoError(v.BindPFlags(fs), test.description)

		err = config.validate()
		if test.expectedErr != "" {
			s.Contains(err.Error(), test.expectedErr, test.description)
		} else {
			s.NoError(err, test.description)
		}
	}
}

func (s *StatsPusher) TestNodeDataURL() {
	u, err := url.ParseRequestURI(s.configData.NodeDataURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.nodeDataURL())
}

func (s *StatsPusher) TestClusterDataURL() {
	u, err := url.ParseRequestURI(s.configData.ClusterDataURL)
	s.Require().NoError(err)
	s.Equal(u, s.config.clusterDataURL())
}

func (s *StatsPusher) TestRequestTimeout() {
	s.EqualValues(s.configData.RequestTimeout, s.config.requestTimeout().String())
}

func (s *StatsPusher) TestDatasetInterval() {
	s.EqualValues(s.configData.DatasetInterval, s.config.datasetInterval().String())
}

func (s *StatsPusher) TestBundleInterval() {
	s.EqualValues(s.configData.BundleInterval, s.config.bundleInterval().String())
}

func (s *StatsPusher) TestNodeInterval() {
	s.EqualValues(s.configData.NodeInterval, s.config.nodeInterval().String())
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
		if err := fs.Set("datasetInterval", configData.DatasetInterval); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("bundleInterval", configData.BundleInterval); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("nodeInterval", configData.NodeInterval); err != nil {
			return nil, nil, nil, configFile, err
		}
		if err := fs.Set("datasetDir", configData.DatasetDir); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	return config, fs, v, configFile, nil
}
