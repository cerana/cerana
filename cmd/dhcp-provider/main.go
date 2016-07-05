package main

import (
	"net"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func dieOnError(msg string, err error) {
	if err != nil {
		log.WithError(err).Fatal(msg)
		os.Exit(1)
	}
}

func defaultNetwork() net.IPNet {
	_, net, _ := net.ParseCIDR("10.0.0.0/8")
	return *net
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})

	v := viper.New()
	f := pflag.NewFlagSet("dhcp-provider", pflag.ExitOnError)
	f.IP("gateway", nil, "default gateway")
	f.Duration("lease-duration", 24*time.Hour, "default lease duration")
	f.IPNet("network", defaultNetwork(), "network to manage dhcp addresses on")
	f.StringSlice("dns-servers", nil, "dns servers")

	config := dhcp.NewConfig(f, v)

	dieOnError("error parsing args", f.Parse(os.Args))
	dieOnError("error loading config", config.LoadConfig())
	dieOnError("error setting up logging", config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError("error creating provider", err)

	d, err := dhcp.New(config, server.Tracker())
	dieOnError("error creating dhcp server", err)

	d.RegisterTasks(server)
	if len(server.RegisteredTasks()) == 0 {
		log.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	dieOnError("dhcp provider encountered and error", server.Start())
	server.StopOnSignal()
}
