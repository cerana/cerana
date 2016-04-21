package clusterconf_test

import (
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/provider"
	clusterconfp "github.com/cerana/cerana/providers/clusterconf"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type clusterConf struct {
	suite.Suite
	config       *clusterconfp.Config
	clusterConf  *clusterconfp.ClusterConf
	flagset      *pflag.FlagSet
	viper        *viper.Viper
	tracker      *acomm.Tracker
	dir          string
	responseHook *url.URL
}

func TestSystemd(t *testing.T) {
	suite.Run(t, new(clusterConf))
}

func (s *clusterConf) SetupSuite() {
	s.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	dir, err := ioutil.TempDir("", "clusterconf-provider-test-")
	s.Require().NoError(err)
	s.dir = dir

	v := viper.New()
	flagset := pflag.NewFlagSet("clusterconf", pflag.PanicOnError)
	config := clusterconfp.NewConfig(flagset, v)
	s.Require().NoError(flagset.Parse([]string{}))
	v.Set("service_name", "clusterconf-provider-test")
	v.Set("socket_dir", s.dir)
	v.Set("coordinator_url", "unix:///tmp/foobar")
	v.Set("log_level", "fatal")
	s.Require().NoError(config.LoadConfig())
	s.Require().NoError(config.SetupLogging())
	s.config = config
	s.flagset = flagset
	s.viper = v

	tracker, err := acomm.NewTracker(filepath.Join(s.dir, "tracker.sock"), nil, nil, 5*time.Second)
	s.Require().NoError(err)
	s.Require().NoError(tracker.Start())
	s.tracker = tracker

	s.clusterConf = clusterconfp.New(config, tracker)
}

func (s *clusterConf) TearDownSuite() {
	_ = os.RemoveAll(s.dir)
}

func (s *clusterConf) TestRegisterTasks() {
	server, err := provider.NewServer(s.config.Config)
	s.Require().NoError(err)

	s.clusterConf.RegisterTasks(server)

	s.True(len(server.RegisteredTasks()) > 0)
}
