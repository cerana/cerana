package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/systemd"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := systemd.NewConfig(nil, nil)
	flag.StringP("unit_file_dir", "d", "", "directory in which to create unit files")
	flag.Parse()

	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := provider.NewServer(config.Config)
	logrusx.DieOnError(err, "new server")
	s, err := systemd.New(config)
	logrusx.DieOnError(err, "new systemd")
	s.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		logrusx.DieOnError(server.Start(), "start server")
		server.StopOnSignal()
	} else {
		logrus.Warn("no registered tasks, exiting")
	}
}
