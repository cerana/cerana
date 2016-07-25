package service_test

import (
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/service"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type Provider struct {
	suite.Suite
	coordinator  *test.Coordinator
	config       *service.Config
	provider     *service.Provider
	tracker      *acomm.Tracker
	flagset      *pflag.FlagSet
	viper        *viper.Viper
	responseHook *url.URL
	systemd      *systemd.MockSystemd
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(Provider))
}

func (s *Provider) SetupSuite() {
	var err error
	s.coordinator, err = test.NewCoordinator("")
	s.Require().NoError(err)

	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := s.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("service", pflag.PanicOnError)
	v.Set("rollback_clone_cmd", "foo/bar")
	v.Set("dataset_clone_dir", "tmp")
	config := service.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.viper = v

	s.tracker, err = acomm.NewTracker(filepath.Join(s.coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(s.tracker.Start())

	s.provider = service.New(config, s.tracker)

	s.systemd = systemd.NewMockSystemd()
	s.coordinator.RegisterProvider(s.systemd)

	s.Require().NoError(s.coordinator.Start())
}

func (s *Provider) TearDownTest() {
	s.systemd.ClearData()
}

func (s *Provider) TearDownSuite() {
	s.coordinator.Stop()
	s.tracker.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
}

func (s *Provider) TestRegisterTasks() {
	server, err := provider.NewServer(s.config.Config)
	s.Require().NoError(err)

	s.provider.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
