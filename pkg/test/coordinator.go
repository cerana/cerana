package test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/cerana/cerana/acomm"
	"github.com/cerana/cerana/coordinator"
	"github.com/cerana/cerana/provider"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Coordinator holds a coordinator server and a provider server with one or
// more registered mock Providers to be used for testing.
type Coordinator struct {
	SocketDir      string
	coordinatorURL string
	coordinator    *coordinator.Server
	providerName   string
	providerServer *provider.Server
}

// NewCoordinator creates a new Coordinator. The coordinator sever will be
// given a temporary socket directory and external port.
func NewCoordinator(baseDir string) (*Coordinator, error) {
	coordinatorName := "testCoordinator"
	socketDir, err := ioutil.TempDir(baseDir, coordinatorName)
	if err != nil {
		return nil, err
	}

	coordinatorViper := viper.New()
	coordinatorViper.Set("service_name", coordinatorName)
	coordinatorViper.Set("socket_dir", socketDir)
	coordinatorViper.Set("external_port", 1024+rand.Intn(65535-1024))
	coordinatorViper.Set("request_timeout", 20)
	coordinatorViper.Set("log_level", "fatal")

	flags := pflag.NewFlagSet(coordinatorName, pflag.ContinueOnError)
	coordinatorConfig := coordinator.NewConfig(flags, coordinatorViper)
	if err = flags.Parse([]string{}); err != nil {
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

	c.providerName = "testProvider"
	providerFlags := pflag.NewFlagSet(c.providerName, pflag.ContinueOnError)
	providerConfig := provider.NewConfig(providerFlags, c.NewProviderViper())
	if err = providerFlags.Parse([]string{}); err != nil {
		return nil, err
	}

	providerServer, err := provider.NewServer(providerConfig)
	if err != nil {
		return nil, err
	}
	c.providerServer = providerServer

	return c, nil
}

// NewProviderViper prepares a basic viper instance for a Provider, setting
// appropriate values corresponding to the coordinator and provider server.
func (c *Coordinator) NewProviderViper() *viper.Viper {
	v := viper.New()
	v.Set("service_name", c.providerName)
	v.Set("socket_dir", c.SocketDir)
	v.Set("coordinator_url", c.coordinatorURL)
	v.Set("request_timeout", 20)
	v.Set("log_level", "fatal")
	return v
}

// ProviderTracker returns the tracker of the provider server.
func (c *Coordinator) ProviderTracker() *acomm.Tracker {
	return c.providerServer.Tracker()
}

// RegisterProvider registers a Provider's tasks with the internal Provider
// server.
func (c *Coordinator) RegisterProvider(p provider.Provider) {
	p.RegisterTasks(c.providerServer)
}

// Start starts the Coordinator and Provider servers.
func (c *Coordinator) Start() error {
	if err := c.coordinator.Start(); err != nil {
		return err
	}
	if err := c.providerServer.Start(); err != nil {
		return err
	}
	return nil
}

// Stop stops the Coordinator and Provider servers.
func (c *Coordinator) Stop() {
	c.providerServer.Stop()
	c.coordinator.Stop()
}

// Cleanup removes the temporary socket directory.
func (c *Coordinator) Cleanup() error {
	return os.RemoveAll(c.SocketDir)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
