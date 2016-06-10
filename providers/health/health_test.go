package health_test

import (
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	healthp "github.com/cerana/cerana/providers/health"
	"github.com/cerana/cerana/providers/systemd"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type health struct {
	suite.Suite
	coordinator  *test.Coordinator
	config       *provider.Config
	tracker      *acomm.Tracker
	viper        *viper.Viper
	health       *healthp.Health
	systemd      *systemd.MockSystemd
	responseHook *url.URL
}

func TestHealth(t *testing.T) {
	suite.Run(t, new(health))
}

func (s *health) SetupSuite() {
	var err error
	s.coordinator, err = test.NewCoordinator("")
	s.Require().NoError(err)

	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := s.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("go-health", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.viper = v

	s.tracker, err = acomm.NewTracker(filepath.Join(s.coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(s.tracker.Start())

	s.health = healthp.New(s.config, s.tracker)

	s.systemd = systemd.NewMockSystemd()
	s.coordinator.RegisterProvider(s.systemd)

	s.Require().NoError(s.coordinator.Start())
}

func (s *health) TearDownSuite() {
	s.coordinator.Stop()
	s.tracker.Stop()
	s.Require().NoError(s.coordinator.Cleanup())
}

func (s *health) TestRegisterTasks() {
	server, err := provider.NewServer(s.config)
	s.Require().NoError(err)

	s.health.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
