package main

import (
	"net"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/dhcp"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func dieOnError(msg string, err error) {
	if err != nil {
		logrus.WithError(err).Fatal("failed to " + msg)
	}
}

func defaultNetwork() net.IPNet {
	_, net, _ := net.ParseCIDR("10.0.0.0/8")
	return *net
}

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	v := viper.New()
	f := pflag.NewFlagSet("dhcp-provider", pflag.ExitOnError)
	f.String("dns-servers", "", "[optional] comma separated list of dns servers ")
	f.IP("gateway", nil, "[optional] default gateway")
	f.Duration("lease-duration", 24*time.Hour, "default lease duration")
	f.IPNet("network", defaultNetwork(), "network to manage dhcp addresses on")

	config := dhcp.NewConfig(f, v)
	dieOnError("parse arguments", f.Parse(os.Args))
	dieOnError("load configuration", config.LoadConfig())
	dieOnError("setup logging", config.SetupLogging())


	server, err := provider.NewServer(config.Config)
	dieOnError("create provider", err)

	d, err := dhcp.New(config, server.Tracker())
	dieOnError("create dhcp server", err)

	d.RegisterTasks(server)
	if len(server.RegisteredTasks()) == 0 {
		logrus.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	dieOnError("successfully run dhcp provider", server.Start())
	server.StopOnSignal()
}
