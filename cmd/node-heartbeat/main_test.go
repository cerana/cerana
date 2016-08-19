package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/tick"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type NodeHeartbeat struct {
	suite.Suite
	config      *tick.Config
	configData  *tick.ConfigData
	configFile  *os.File
	tracker     *acomm.Tracker
	coordinator *test.Coordinator
	clusterConf *clusterconf.MockClusterConf
	metrics     *metrics.MockMetrics
}

func TestNodeHeartbeat(t *testing.T) {
	suite.Run(t, new(NodeHeartbeat))
}

func (s *NodeHeartbeat) SetupSuite() {
	noError := s.Require().NoError

	logrus.SetLevel(logrus.FatalLevel)

	// Setup mock coordinator
	var err error
	s.coordinator, err = test.NewCoordinator("")
	noError(err)

	nodeDataURL := s.coordinator.NewProviderViper().GetString("coordinator_url")
	s.configData = &tick.ConfigData{
		NodeDataURL:       nodeDataURL,
		ClusterDataURL:    nodeDataURL,
		LogLevel:          "fatal",
		RequestTimeout:    "5s",
		TickInterval:      "4s",
		TickRetryInterval: "4s",
	}

	s.config, _, _, s.configFile, err = newTestConfig(false, true, s.configData)
	noError(err, "failed to create config")
	noError(s.config.LoadConfig(), "failed to load config")

	tracker, err := acomm.NewTracker("", nil, nil, s.config.RequestTimeout())
	noError(err)
	s.tracker = tracker
	noError(s.tracker.Start())

	// Setup mock providers
	s.setupClusterConf()
	s.setupMetrics()

	noError(s.coordinator.Start())
}

func (s *NodeHeartbeat) setupClusterConf() {
	s.clusterConf = clusterconf.NewMockClusterConf()
	s.coordinator.RegisterProvider(s.clusterConf)
}

func (s *NodeHeartbeat) TearDownSuite() {
	s.coordinator.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
	_ = os.Remove(s.configFile.Name())
	s.tracker.Stop()
}

func (s *NodeHeartbeat) setupMetrics() {
	s.metrics = metrics.NewMockMetrics()
	s.coordinator.RegisterProvider(s.metrics)
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
		configFile, err = ioutil.TempFile("", "nodeHeartbeat-")
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
