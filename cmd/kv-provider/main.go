package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/kv"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := kv.NewConfig(nil, nil)
	flag.StringP("address", "a", "", "kv address (leave blank for default)")
	flag.Parse()

	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := provider.NewServer(config.Config)
	logrusx.DieOnError(err, "new server")

	k, err := kv.New(config, server.Tracker())
	logrusx.DieOnError(err, "new kv")
	k.RegisterTasks(server)

	if len(server.RegisteredTasks()) == 0 {
		logrus.Warn("no registered tasks, exiting")
		os.Exit(1)
	}
	logrusx.DieOnError(server.Start(), "start server")
	server.StopOnSignal()
}
