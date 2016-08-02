package main

import (
	"os"

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

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)
	s, err := systemd.New(config)
	dieOnError(err)
	s.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		dieOnError(server.Start())
		server.StopOnSignal()
	} else {
		logrus.Warn("no registered tasks, exiting")
	}
}

func dieOnError(err error) {
	if err != nil {
		logrus.Fatal("encountered an error during startup, error:", err)
		os.Exit(1)
	}
}
