package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/provider/examples/simple"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := provider.NewConfig(nil, nil)

	flag.StringP("foobar", "f", "", "path to config file")

	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config)
	dieOnError(err)
	s := simple.NewSimple(config, server.Tracker())
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
		logrus.Fatal("encountered an error during startup")
	}
}
