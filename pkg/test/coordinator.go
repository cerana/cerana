package test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/cerana/cerana/coordinator"
	"github.com/cerana/cerana/provider"
	"github.com/pborman/uuid"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type Coordinator struct {
	SocketDir      string
	coordinatorURL string
	coordinator    *coordinator.Server
	providerName   string
	providerServer *provider.Server
}

func NewCoordinator(baseDir string) (*Coordinator, error) {
	coordinatorName := uuid.New()
	socketDir, err := ioutil.TempDir(baseDir, coordinatorName)
	if err != nil {
		return nil, err
	}

	coordinatorViper := viper.New()
	coordinatorViper.Set("service_name", coordinatorName)
	coordinatorViper.Set("socket_dir", socketDir)
	coordinatorViper.Set("external_port", 10000+rand.Intn(10000))
	coordinatorViper.Set("request_timeout", 20)
	coordinatorViper.Set("log_level", "fatal")

	flags := pflag.NewFlagSet(coordinatorName, pflag.ContinueOnError)
	coordinatorConfig := coordinator.NewConfig(flags, coordinatorViper)
	if err := flags.Parse([]string{}); err != nil {
		return nil, err
	}
	_ = coordinatorConfig.SetupLogging()

	coordinatorServer, err := coordinator.NewServer(coordinatorConfig)
	if err != nil {
		return nil, err
	}

	coordinatorSocket := "unix://" + filepath.Join(
		coordinatorConfig.SocketDir(),
		"coordinator",
		coordinatorConfig.ServiceName()+".sock")

	c := &Coordinator{
		SocketDir:      socketDir,
		coordinatorURL: coordinatorSocket,
		coordinator:    coordinatorServer,
	}

	c.providerName = uuid.New()
	providerFlags := pflag.NewFlagSet(c.providerName, pflag.ContinueOnError)
	providerConfig := provider.NewConfig(providerFlags, c.NewProviderViper())
	if err := providerFlags.Parse([]string{}); err != nil {
		return nil, err
	}

	providerServer, err := provider.NewServer(providerConfig)
	if err != nil {
		return nil, err
	}
	c.providerServer = providerServer

	return c, nil
}

func (c *Coordinator) NewProviderViper() *viper.Viper {
	v := viper.New()
	v.Set("service_name", c.providerName)
	v.Set("socket_dir", c.SocketDir)
	v.Set("coordinator_url", c.coordinatorURL)
	v.Set("request_timeout", 20)
	v.Set("log_level", "fatal")
	return v
}

func (c *Coordinator) RegisterProvider(p provider.Provider) {
	p.RegisterTasks(*c.providerServer)
}

func (c *Coordinator) Start() error {
	if err := c.coordinator.Start(); err != nil {
		return err
	}
	if err := c.providerServer.Start(); err != nil {
		return err
	}
	return nil
}

func (c *Coordinator) Stop() {
	c.providerServer.Stop()
	c.coordinator.Stop()
}

func (c *Coordinator) Cleanup() error {
	return os.RemoveAll(c.SocketDir)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
