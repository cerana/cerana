package health_test

import (
	"net/url"
	"testing"

	"github.com/cerana/cerana/provider"
	healthp "github.com/cerana/cerana/providers/health"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type health struct {
	suite.Suite
	health       *healthp.Health
	config       *provider.Config
	responseHook *url.URL
}

func TestHealth(t *testing.T) {
	suite.Run(t, new(health))
}

func (s *health) SetupSuite() {
	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	s.health = &healthp.Health{}

	v := viper.New()
	flagset := pflag.NewFlagSet("go-health", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "health-provider-test")
	v.Set("socket_dir", "/tmp")
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
}

func (s *health) TestRegisterTasks() {
	server, err := provider.NewServer(s.config)
	s.Require().NoError(err)

	s.health.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
