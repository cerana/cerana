package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/clusterconf"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := clusterconf.NewConfig(nil, nil)
	flag.DurationP("dataset_ttl", "d", time.Minute, "ttl for dataset usage heartbeats")
	flag.DurationP("bundle_ttl", "b", time.Minute, "ttl for bundle usage heartbeats")
	flag.DurationP("node_ttl", "o", time.Minute, "ttl for node heartbeats")
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)
	c := clusterconf.New(config, server.Tracker())
	dieOnError(err)
	c.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		dieOnError(server.Start())
		server.StopOnSignal()
	} else {
		logrus.Warn("no registered tasks, exiting")
	}
}

func dieOnError(err error) {
	if err != nil {
		logrus.Fatal("encountered an error during startup")
	}
}
