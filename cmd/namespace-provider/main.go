package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/namespace"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := provider.NewConfig(nil, nil)
	flag.Parse()

	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := provider.NewServer(config)
	logrusx.DieOnError(err, "new server")
	n := namespace.New(config, server.Tracker())
	n.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		logrusx.DieOnError(server.Start(), "start server")
		server.StopOnSignal()
	} else {
		logrus.Warn("no registered tasks, exiting")
	}
}
