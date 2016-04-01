package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/cerana/cerana/provider"
	"github.com/cerana/cerana/providers/zfs"
	logx "github.com/cerana/cerana/pkg/logrusx"
	flag "github.com/spf13/pflag"
)

func main() {
	log.SetFormatter(&logx.MistifyFormatter{})

	config := provider.NewConfig(nil, nil)
	flag.Parse()

	dieOnError(config.LoadConfig())
	dieOnError(config.SetupLogging())

	server, err := provider.NewServer(config)
	dieOnError(err)
	z := zfs.New(config, server.Tracker())
	z.RegisterTasks(server)

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
