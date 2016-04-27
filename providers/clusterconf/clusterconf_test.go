package clusterconf_test

import (
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/mistifyio/lochness/pkg/kv"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type clusterConf struct {
	suite.Suite
	coordinator  *test.Coordinator
	config       *clusterconf.Config
	clusterConf  *clusterconf.ClusterConf
	viper        *viper.Viper
	responseHook *url.URL
	kvp          *KVP
}

func TestClusterConf(t *testing.T) {
	suite.Run(t, new(clusterConf))
}

func (s *clusterConf) SetupSuite() {
	var err error
	s.coordinator, err = test.NewCoordinator("")
	s.Require().NoError(err)

	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := s.coordinator.NewProviderViper()
	v.Set("dataset_ttl", time.Minute)
	v.Set("bundle_ttl", time.Minute)
	v.Set("node_ttl", time.Minute)
	flagset := pflag.NewFlagSet("clusterconf", pflag.PanicOnError)
	config := clusterconf.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.viper = v

	tracker, err := acomm.NewTracker(filepath.Join(s.coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(tracker.Start())

	s.clusterConf = clusterconf.New(config, tracker)

	v = s.coordinator.NewProviderViper()
	flagset = pflag.NewFlagSet("clusterconf", pflag.PanicOnError)
	config = clusterconf.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.kvp = NewKVP(config.Config)
	s.coordinator.RegisterProvider(s.kvp)

	s.Require().NoError(s.coordinator.Start())
}

func (s *clusterConf) TearDownTest() {
	s.kvp.Data = make(map[string]kv.Value)
}

func (s *clusterConf) TearDownSuite() {
	s.coordinator.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
}

func (s *clusterConf) TestRegisterTasks() {
	server, err := provider.NewServer(s.config.Config)
	s.Require().NoError(err)

	s.clusterConf.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
