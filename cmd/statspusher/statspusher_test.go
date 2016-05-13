package main

import (
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/suite"
)

type StatsPusher struct {
	suite.Suite
	config      *config
	configData  *ConfigData
	configFile  *os.File
	statsPusher *statsPusher
	tracker     *acomm.Tracker
	coordinator *test.Coordinator
	systemd     *systemd.MockSystemd
	zfs         *zfs.MockZFS
	clusterConf *clusterconf.MockClusterConf
}

func TestStatsPusher(t *testing.T) {
	suite.Run(t, new(StatsPusher))
}

func (s *StatsPusher) SetupSuite() {
	noError := s.Require().NoError

	logrus.SetLevel(logrus.FatalLevel)

	// Setup mock coordinator
	var err error
	s.coordinator, err = test.NewCoordinator("")
	noError(err)

	coordinatorURL := s.coordinator.NewProviderViper().GetString("coordinator_url")
	s.configData = &ConfigData{
		CoordinatorURL: coordinatorURL,
		HeartbeatURL:   coordinatorURL,
		LogLevel:       "fatal",
		RequestTimeout: 5,
		DatasetTTL:     4,
		BundleTTL:      3,
		NodeTTL:        2,
	}

	s.config, _, _, s.configFile, err = newTestConfig(false, true, s.configData)
	noError(err, "failed to create config")
	noError(s.config.loadConfig(), "failed to load config")

	s.statsPusher, err = newStatsPusher(s.config)
	noError(err)
	noError(s.statsPusher.tracker.Start())

	// Setup mock providers
	noError(err)

	s.setupSystemd()
	s.setupZFS()
	s.setupClusterConf()

	noError(s.coordinator.Start())
}

func (s *StatsPusher) setupClusterConf() {
	s.clusterConf = clusterconf.NewMockClusterConf()
	s.coordinator.RegisterProvider(s.clusterConf)
}

func (s *StatsPusher) setupZFS() {
	v := s.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("zfs", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.zfs = zfs.NewMockZFS(config, s.coordinator.ProviderTracker())
	s.coordinator.RegisterProvider(s.zfs)
}

func (s *StatsPusher) setupSystemd() {
	s.systemd = systemd.NewMockSystemd()
	s.coordinator.RegisterProvider(s.systemd)
}

func (s *StatsPusher) TearDownSuite() {
	s.coordinator.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
	_ = os.Remove(s.configFile.Name())
	s.statsPusher.tracker.Stop()
}