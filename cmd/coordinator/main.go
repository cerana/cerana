package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/coordinator"
	"github.com/cerana/cerana/pkg/logrusx"
	flag "github.com/spf13/pflag"
)

func main() {
	logrus.SetFormatter(&logrusx.JSONFormatter{})

	config := coordinator.NewConfig(nil, nil)
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := coordinator.NewServer(config)
	dieOnError(err)

	dieOnError(server.Start())
	server.StopOnSignal()
}

func dieOnError(err error) {
	if err != nil {
		logrus.Fatal("encountered an error during startup, error:", err)
		os.Exit(1)
	}
}
