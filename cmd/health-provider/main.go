package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	logx "github.com/cerana/cerana/pkg/logrusx"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/health"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.JSONFormatter{})

	config := provider.NewConfig(nil, nil)
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config)
	dieOnError(err)
	h := health.New(config, server.Tracker())
	h.RegisterTasks(server)

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
