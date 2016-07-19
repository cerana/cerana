package main

import (
	log "github.com/Sirupsen/logrus"
	logx "github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/datatrade"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.JSONFormatter{})

	config := datatrade.NewConfig(nil, nil)
	flag.UintP("node_coordinator_port", "o", 0, "node coordinator external port")
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config.Config)
	dieOnError(err)
	d := datatrade.New(config, server.Tracker())
	dieOnError(err)
	d.RegisterTasks(server)

	if len(server.RegisteredTasks()) != 0 {
		dieOnError(server.Start())
		server.StopOnSignal()
	} else {
		log.Warn("no registered tasks, exiting")
	}
}

func dieOnError(err error) {
	if err != nil {
		log.Fatal("encountered an error during startup")
	}
}
