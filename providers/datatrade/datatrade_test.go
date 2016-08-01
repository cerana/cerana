package datatrade_test

import (
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/pkg/test"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	"github.com/cerana/cerana/providers/datatrade"
	"github.com/cerana/cerana/providers/zfs"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type Provider struct {
	suite.Suite
	nodeCoordinator *test.Coordinator
	coordinator     *test.Coordinator
	config          *datatrade.Config
	provider        *datatrade.Provider
	tracker         *acomm.Tracker
	flagset         *pflag.FlagSet
	viper           *viper.Viper
	responseHook    *url.URL
	zfs             *zfs.MockZFS
	clusterConf     *clusterconf.MockClusterConf
}

func TestProvider(t *testing.T) {
	suite.Run(t, new(Provider))
}

func (p *Provider) SetupSuite() {
	var err error
	p.nodeCoordinator, err = test.NewCoordinator("")
	p.Require().NoError(err)

	p.coordinator, err = test.NewCoordinator("")
	p.Require().NoError(err)

	p.responseHook, _ = url.ParseRequestURI("unix:///tmp/foobar")

	v := p.coordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("datatrade", pflag.PanicOnError)
	v.Set("dataset_dir", "foobar")
	v.Set("node_coordinator_port", uint(p.nodeCoordinator.HTTPPort))
	config := datatrade.NewConfig(flagset, v)
	p.Require().NoError(flagset.Parse([]string{}))
	p.Require().NoError(config.LoadConfig())
	p.Require().NoError(config.SetupLogging())
	p.config = config
	p.viper = v

	p.tracker, err = acomm.NewTracker(filepath.Join(p.coordinator.SocketDir, "tracker.sock"), nil, nil, 5*time.Second)
	p.Require().NoError(err)
	p.Require().NoError(p.tracker.Start())

	p.provider = datatrade.New(p.config, p.tracker)

	p.setupClusterConf()
	p.setupZFS()

	p.Require().NoError(p.nodeCoordinator.Start())
	p.Require().NoError(p.coordinator.Start())
}

func (p *Provider) setupClusterConf() {
	p.clusterConf = clusterconf.NewMockClusterConf()
	p.coordinator.RegisterProvider(p.clusterConf)
}

func (p *Provider) setupZFS() {
	v := p.nodeCoordinator.NewProviderViper()
	flagset := pflag.NewFlagSet("zfs", pflag.PanicOnError)
	config := provider.NewConfig(flagset, v)
	p.Require().NoError(flagset.Parse([]string{}))
	p.Require().NoError(config.LoadConfig())
	p.zfs = zfs.NewMockZFS(config, p.nodeCoordinator.ProviderTracker())
	p.nodeCoordinator.RegisterProvider(p.zfs)
}

func (p *Provider) TearDownSuite() {
	p.nodeCoordinator.Stop()
	p.coordinator.Stop()
	p.tracker.Stop()
	p.Require().NoError(p.nodeCoordinator.Cleanup())
	p.Require().NoError(p.coordinator.Cleanup())
}

func (p *Provider) TestRegisterTasks() {
	server, err := provider.NewServer(p.config.Config)
	p.Require().NoError(err)

	p.clusterConf.RegisterTasks(server)

	p.True(len(server.RegisteredTasks()) > 0)
}
