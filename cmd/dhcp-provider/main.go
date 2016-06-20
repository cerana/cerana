package main

import (
	"net"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/spf13/pflag"
)

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup, error:", err)
		os.Exit(1)
	}
}

func defaultNetwork() net.IPNet {
	_, net, _ := net.ParseCIDR("10.0.0.0/8")
	return *net
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	config := dhcp.NewConfig(nil, nil)
	pflag.IP("gateway", nil, "default gateway")
	pflag.Duration("lease-duration", 24*time.Hour, "default lease duration")
	pflag.IPNet("network", defaultNetwork(), "network to manage dhcp addresses on")
	pflag.StringSlice("dns-servers", nil, "dns servers")
	pflag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)

	d, err := dhcp.New(config, server.Tracker())
	dieOnError(err)

	d.RegisterTasks(server)
	if len(server.RegisteredTasks()) == 0 {
		log.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	dieOnError(server.Start())
	server.StopOnSignal()
}
