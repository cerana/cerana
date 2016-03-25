package metrics_test

import (
	"testing"

	"github.com/mistifyio/mistify/provider"
	metricsp "github.com/mistifyio/mistify/providers/metrics"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type metrics struct {
	suite.Suite
	metrics *metricsp.Metrics
	config  *provider.Config
}

func TestMetrics(t *testing.T) {
	suite.Run(t, new(metrics))
}

func (s *metrics) SetupSuite() {
	s.metrics = &metricsp.Metrics{}

	v := viper.New()
	flagset := pflag.NewFlagSet("go-metrics", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "metrics-provider-test")
	v.Set("socket_dir", "/tmp")
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
}

func (s *metrics) TestRegisterTasks() {
	server, err := provider.NewServer(s.config)
	s.Require().NoError(err)

	s.metrics.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
