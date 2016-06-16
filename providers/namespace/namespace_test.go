package namespace_test

import (
	"net/url"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/namespace"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type Namespace struct {
	suite.Suite
	responseHook *url.URL
	namespace    *namespace.Namespace
	config       *provider.Config
}

func TestNamespace(t *testing.T) {
	suite.Run(t, new(Namespace))
}

func (s *Namespace) SetupSuite() {
	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := viper.New()
	flagset := pflag.NewFlagSet("namespace", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "namespace-provider-test")
	v.Set("socket_dir", "/tmp")
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config

	s.namespace = namespace.New(config, nil)
}

func (s *Namespace) TestRegisterTasks() {
	server, err := provider.NewServer(s.config)
	s.Require().NoError(err)

	s.namespace.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}

func newProc() (*os.Process, error) {
	cmd := exec.Command("unshare", "-U", "sleep", "60")
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	time.Sleep(time.Millisecond)
	return cmd.Process, nil
}
