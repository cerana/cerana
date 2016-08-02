package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/service"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := service.NewConfig(nil, nil)
	flag.StringP("rollback_clone_cmd", "r", "/run/current-system/sw/bin/rollback_clone", "full path to dataset clone/rollback tool")
	flag.StringP("dataset_clone_dir", "d", "data/running-clones", "destination for dataset clones used by running services")
	flag.Parse()

	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := provider.NewServer(config.Config)
	logrusx.DieOnError(err, "new server")
	s := service.New(config, server.Tracker())
	s.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		logrusx.DieOnError(server.Start(), "start server")
		server.StopOnSignal()
	} else {
		logrus.Warn("no registered tasks, exiting")
	}
}
