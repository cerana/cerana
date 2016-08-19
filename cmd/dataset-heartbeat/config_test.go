package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"

	"github.com/cerana/cerana/tick"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func (s *DatasetHeartbeat) TestValidate() {
	u := "unix:///tmp/foobar"
	tests := []struct {
		datasetPrefix string
		expectedErr   string
	}{
		{"foobar", ""},
		{"", "missing datasetPrefix"},
	}

	for _, test := range tests {
		configData := &ConfigData{
			ConfigData: tick.ConfigData{
				NodeDataURL:       u,
				ClusterDataURL:    u,
				RequestTimeout:    "5s",
				TickInterval:      "4s",
				TickRetryInterval: "3s",
			},
			DatasetPrefix: test.datasetPrefix,
		}

		config, fs, v, _, err := newTestConfig(true, false, configData)
		if !s.NoError(err, test.datasetPrefix) {
			continue
		}
		// Bind here to avoid the need for Load
		s.Require().NoError(v.BindPFlags(fs), test.datasetPrefix)

		err = config.Validate()
		if test.expectedErr != "" {
			s.Contains(err.Error(), test.expectedErr, test.datasetPrefix)
		} else {
			s.NoError(err, test.datasetPrefix)
		}
	}
}

func (s *DatasetHeartbeat) TestDatasetPrefix() {
	s.EqualValues(s.configData.DatasetPrefix, s.config.DatasetPrefix())
}

func newTestConfig(setFlags, writeConfig bool, configData *ConfigData) (*Config, *pflag.FlagSet, *viper.Viper, *os.File, error) {
	fs := pflag.NewFlagSet(uuid.New(), pflag.ExitOnError)
	v := viper.New()
	v.SetConfigType("json")
	config := NewConfig(fs, v)
	if config == nil {
		return nil, nil, nil, nil, errors.New("failed to return a config")
	}

	var configFile *os.File
	if writeConfig {
		var err error
		configFile, err = ioutil.TempFile("", "datasetHeartbeat-")
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
		if err := fs.Set("datasetPrefix", configData.DatasetPrefix); err != nil {
			return nil, nil, nil, configFile, err
		}
	}

	return config, fs, v, configFile, nil
}
