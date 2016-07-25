package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	logx "github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/service"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.JSONFormatter{})

	config := service.NewConfig(nil, nil)
	flag.StringP("dataset_clone_bin", "b", "/run/current-system/sw/bin/rollback_clone", "full path to dataset clone/rollback binary")
	flag.StringP("dataset_clone_path", "p", "data/running-clones", "path for dataset clones used by running services")
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)
	s := service.New(config, server.Tracker())
	s.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		dieOnError(server.Start())
		server.StopOnSignal()
	} else {
		log.Warn("no registered tasks, exiting")
	}
}

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup, error:", err)
		os.Exit(1)
	}
}
