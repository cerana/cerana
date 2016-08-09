package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/coordinator"
	"github.com/cerana/cerana/pkg/logrusx"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := coordinator.NewConfig(nil, nil)
	flag.Parse()

	logrusx.DieOnError(config.LoadConfig(), "load config")
	logrusx.DieOnError(config.SetupLogging(), "setup logging")

	server, err := coordinator.NewServer(config)
	logrusx.DieOnError(err, "new server")

	logrusx.DieOnError(server.Start(), "start server")
	server.StopOnSignal()
}
