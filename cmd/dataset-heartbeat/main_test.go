package main

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/metrics"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/cerana/cerana/tick"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

type DatasetHeartbeat struct {
	suite.Suite
	config      *Config
	configData  *ConfigData
	configFile  *os.File
	tracker     *acomm.Tracker
	coordinator *test.Coordinator
	zfs         *zfs.MockZFS
	clusterConf *clusterconf.MockClusterConf
	metrics     *metrics.MockMetrics
}

func TestDatasetHeartbeat(t *testing.T) {
	suite.Run(t, new(DatasetHeartbeat))
}

func (s *DatasetHeartbeat) SetupSuite() {
	noError := s.Require().NoError

	logrus.SetLevel(logrus.FatalLevel)

	// Setup mock coordinator
	var err error
	s.coordinator, err = test.NewCoordinator("")
	noError(err)

	nodeDataURL := s.coordinator.NewProviderViper().GetString("coordinator_url")
	s.configData = &ConfigData{
		ConfigData: tick.ConfigData{
			NodeDataURL:       nodeDataURL,
			ClusterDataURL:    nodeDataURL,
			LogLevel:          "fatal",
			RequestTimeout:    "5s",
			TickInterval:      "4s",
			TickRetryInterval: "4s",
		},
		DatasetPrefix: "foobar",
	}

	s.config, _, _, s.configFile, err = newTestConfig(false, true, s.configData)
	noError(err, "failed to create config")
	noError(s.config.LoadConfig(), "failed to load config")

	tracker, err := acomm.NewTracker("", nil, nil, s.config.RequestTimeout())
	noError(err)
	s.tracker = tracker
	noError(s.tracker.Start())

	// Setup mock providers
	s.setupZFS()
	s.setupClusterConf()
	s.setupMetrics()

	noError(s.coordinator.Start())
}

func (s *DatasetHeartbeat) setupClusterConf() {
	s.clusterConf = clusterconf.NewMockClusterConf()
	s.coordinator.RegisterProvider(s.clusterConf)
}

func (s *DatasetHeartbeat) setupZFS() {
	v := s.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("zfs", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.zfs = zfs.NewMockZFS(config, s.coordinator.ProviderTracker())
	s.coordinator.RegisterProvider(s.zfs)
}

func (s *DatasetHeartbeat) TearDownSuite() {
	s.coordinator.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
	_ = os.Remove(s.configFile.Name())
	s.tracker.Stop()
}

func (s *DatasetHeartbeat) setupMetrics() {
	s.metrics = metrics.NewMockMetrics()
	s.coordinator.RegisterProvider(s.metrics)
}
